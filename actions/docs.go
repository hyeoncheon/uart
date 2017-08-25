package actions

//! WIP
//* Use Belonging Interface

import (
	"net/http"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/hyeoncheon/uart/models"
	"github.com/markbates/inflect"
	"github.com/markbates/pop"
	"github.com/pkg/errors"
)

// DocsResource is the resource for the doc model
type DocsResource struct {
	buffalo.Resource
}

// List gets all Docs. GET /docs
func (v DocsResource) List(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	docs := &models.Docs{}
	q := tx.Q()
	if false == c.Value("member_is_admin").(bool) {
		q = q.Where("is_published = ?", true)
	}
	err := q.Order("category, subject").All(docs)
	if err != nil {
		return errors.WithStack(err)
	}

	cat := ""
	sub := ""
	for i := 0; i < len(*docs); i++ {
		doc := &(*docs)[i]
		if doc.Category != cat {
			cat = doc.Category
			doc.NewCategory = doc.Category
			sub = ""
		}
		if doc.Subject != sub {
			sub = doc.Subject
			doc.NewSubject = doc.Subject
		}
	}

	c.Set("docs", docs)
	return c.Render(200, r.HTML("docs/index.html"))
}

// Show gets the data for one Doc. GET /docs/{doc_id}
func (v DocsResource) Show(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	doc := &models.Doc{}
	err := tx.Find(doc, c.Param("doc_id"))
	if err != nil {
		err = tx.Where("slug = ?", c.Param("doc_id")).First(doc)
		if err != nil {
			c.Flash().Add("danger", t(c, "cannot.found.documentation"))
			me := currentMember(c)
			mLogErr(c, MsgFacSecu, "invalid access: docs.show by %v", me)
			return c.Redirect(http.StatusFound, "/docs")
		}
	}
	c.Set("doc", doc)
	return c.Render(200, r.HTML("docs/show.html"))
}

// New renders the formular for creating a new Doc. GET /docs/new
// ADMIN PROTECTED
func (v DocsResource) New(c buffalo.Context) error {
	c.Set("doc", &models.Doc{})
	c.Set("theme", "default")
	return c.Render(200, r.HTML("docs/new.html"))
}

// Create adds a Doc to the DB. POST /docs
// ADMIN PROTECTED
func (v DocsResource) Create(c buffalo.Context) error {
	doc := &models.Doc{}
	err := c.Bind(doc)
	if err != nil {
		return errors.WithStack(err)
	}

	dumme := dummyMember(c)
	doc.MemberID = dumme.ID
	doc.Slug = inflect.Dasherize(strings.ToLower(doc.Title))

	tx := c.Value("tx").(*pop.Connection)
	verrs, err := tx.ValidateAndCreate(doc)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("doc", doc)
		c.Set("errors", verrs)
		c.Set("theme", "default")
		return c.Render(422, r.HTML("docs/new.html"))
	}
	c.Flash().Add("success", "Doc was created successfully")
	return c.Redirect(302, "/docs/%s", doc.ID)
}

// Edit renders a edit formular for a doc. GET /docs/{doc_id}/edit
// ADMIN PROTECTED
func (v DocsResource) Edit(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	doc := &models.Doc{}
	err := models.FindMyOwn(tx.Q(), dummyMember(c), doc, c.Param("doc_id"))
	if err != nil {
		c.Flash().Add("danger", t(c, "you.have.no.right.for.this.doc"))
		me := currentMember(c)
		mLogErr(c, MsgFacSecu, "access violation: docs.edit by %v", me)
		return c.Redirect(http.StatusFound, "/docs")
	}
	c.Set("doc", doc)
	c.Set("theme", "default")
	return c.Render(200, r.HTML("docs/edit.html"))
}

// Update changes a doc in the DB. PUT /docs/{doc_id}
// ADMIN PROTECTED
func (v DocsResource) Update(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	doc := &models.Doc{}
	err := models.FindMyOwn(tx.Q(), dummyMember(c), doc, c.Param("doc_id"))
	if err != nil {
		c.Flash().Add("danger", t(c, "you.have.no.right.for.this.doc"))
		me := currentMember(c)
		mLogErr(c, MsgFacSecu, "access violation: docs.update by %v", me)
		return c.Redirect(http.StatusFound, "/docs")
	}
	err = c.Bind(doc)
	if err != nil {
		return errors.WithStack(err)
	}

	doc.Slug = inflect.Dasherize(strings.ToLower(doc.Title))

	verrs, err := tx.ValidateAndUpdate(doc)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("doc", doc)
		c.Set("errors", verrs)
		c.Set("theme", "default")
		return c.Render(422, r.HTML("docs/edit.html"))
	}
	c.Flash().Add("success", "Doc was updated successfully")
	return c.Redirect(302, "/docs/%s", doc.ID)
}

// Destroy deletes a doc from the DB. DELETE /docs/{doc_id}
// ADMIN PROTECTED
func (v DocsResource) Destroy(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	doc := &models.Doc{}
	err := models.FindMyOwn(tx.Q(), dummyMember(c), doc, c.Param("doc_id"))
	if err != nil {
		c.Flash().Add("danger", t(c, "you.have.no.right.for.this.doc"))
		me := currentMember(c)
		mLogErr(c, MsgFacSecu, "access violation: docs.destroy by %v", me)
		return c.Redirect(http.StatusFound, "/docs")
	}
	err = tx.Destroy(doc)
	if err != nil {
		return errors.WithStack(err)
	}
	c.Flash().Add("success", "Doc was destroyed successfully")
	return c.Redirect(302, "/docs")
}

// Publish marks the document as published. GET /docs/{doc_id}/publish
func (v DocsResource) Publish(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	doc := &models.Doc{}
	err := models.FindMyOwn(tx.Q(), dummyMember(c), doc, c.Param("doc_id"))
	if err != nil {
		c.Flash().Add("danger", t(c, "you.have.no.right.for.this.doc"))
		me := currentMember(c)
		mLogErr(c, MsgFacSecu, "access violation: docs.publish by %v", me)
		return c.Redirect(http.StatusFound, "/docs")
	}

	doc.IsPublished = true

	if err := tx.Save(doc); err != nil {
		c.Flash().Add("danger", t(c, "cannot.publish.the.document"))
		me := currentMember(c)
		mLogErr(c, MsgFacSecu, "publication error: docs.destroy by %v", me)
		return c.Redirect(http.StatusFound, "/docs")
	}
	c.Flash().Add("success", "Doc was updated successfully")
	return c.Redirect(302, "/docs/%s", doc.ID)
}
