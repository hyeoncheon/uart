package actions_test

//* Test coverage: 100%
// testing by normal creation by admin, publishing, and invalid access.

import (
	"net/http"

	"github.com/hyeoncheon/uart/models"
	uuid "github.com/satori/go.uuid"
)

func (as *ActionSuite) Test_DocsResource_A_CreateAndCheck() {
	as.setupMembers()

	doc := &models.Doc{
		Type:     "Lyrics",
		Category: "Music",
		Subject:  "Rock",
		Title:    "I Want Out",
		Content:  "I want out! to live my life alone!",
	}

	as.loginAs(admin) //! login as admin

	// New()
	res := as.HTML("/docs/new").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "form action=")

	// Create()
	res = as.HTML("/docs").Post(doc)
	as.Equal(http.StatusSeeOther, res.Code)
	as.Contains(res.HeaderMap.Get("Location"), "/docs/")
	as.Equal(uuid.Nil, doc.ID)

	err := as.DB.First(doc) // get doc instance
	as.NoError(err)
	as.NotEqual(uuid.Nil, doc.ID)

	// Edit()
	res = as.HTML("/docs/%v/edit", doc.ID).Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "form action=")
	as.Contains(res.Body.String(), doc.Title)

	// Update()
	doc.Title = "I Want Out (Helloween)"
	res = as.HTML("/docs/%v", doc.ID).Put(doc)
	as.Equal(http.StatusSeeOther, res.Code)
	as.Contains(res.HeaderMap.Get("Location"), "/docs/")

	// List()
	res = as.HTML("/docs").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), doc.Title)
	// Show() with id
	res = as.HTML("/docs/%v", doc.ID).Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), doc.Content)
	// Show() with updated slug
	as.DB.Reload(doc)
	res = as.HTML("/docs/%v", doc.Slug).Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), doc.Title)

	// Destroy()
	res = as.HTML("/docs/%v", doc.ID).Delete()
	as.Equal(http.StatusSeeOther, res.Code)
	as.Equal("/docs", res.HeaderMap.Get("Location"))
}

func (as *ActionSuite) Test_DocsResource_B_Publishing() {
	as.setupMembers()

	doc := &models.Doc{
		Type:     "Lyrics",
		Category: "Music",
		Subject:  "Rock",
		Title:    "I Want Out",
		Content:  "I want out! to live my life alone!",
	}

	as.loginAs(admin) //! login as admin

	// Create()
	res := as.HTML("/docs").Post(doc)
	as.Equal(http.StatusSeeOther, res.Code)
	as.Contains(res.HeaderMap.Get("Location"), "/docs/")
	as.Equal(uuid.Nil, doc.ID)

	err := as.DB.First(doc) // get doc instance
	as.NoError(err)
	as.NotEqual(uuid.Nil, doc.ID)

	as.loginAs(other) //! login as other and cannot see
	// List()
	res = as.HTML("/docs").Get()
	as.Equal(http.StatusOK, res.Code)
	as.NotContains(res.Body.String(), doc.Title)
	// Show() by id
	res = as.HTML("/docs/%v", doc.ID).Get()
	as.Equal(http.StatusNotFound, res.Code)
	// Show() by slug
	res = as.HTML("/docs/%v", doc.Slug).Get()
	as.Equal(http.StatusNotFound, res.Code)

	as.loginAs(admin) //! login as admin, for publishing
	// Publish()
	res = as.HTML("/docs/%v/publish", doc.ID).Get()
	as.Equal(http.StatusSeeOther, res.Code)
	as.Contains(res.HeaderMap.Get("Location"), "/docs/")

	as.loginAs(other) //! login as other and can see
	// List()
	res = as.HTML("/docs").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), doc.Title)
	// Show() by id
	res = as.HTML("/docs/%v", doc.ID).Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), doc.Content)
	// Show() by slug
	res = as.HTML("/docs/%v", doc.Slug).Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), doc.Content)
}

func (as *ActionSuite) Test_DocsResource_C_InvalidAccess() {
	as.setupMembers()

	doc := &models.Doc{
		Type:     "Lyrics",
		Category: "Music",
		Subject:  "Rock",
		Title:    "I Want Out",
		Content:  "I want out! to live my life alone!",
	}

	as.loginAs(other) //! invalid access by other

	// New() as other, ADMIN PROTECTED
	res := as.HTML("/docs/new").Get()
	as.Equal(http.StatusFound, res.Code)
	as.Equal("/", res.HeaderMap.Get("Location"))

	// Create() as other, ADMIN PROTECTED
	res = as.HTML("/docs").Post(doc)
	as.Equal(http.StatusFound, res.Code)
	as.Equal("/", res.HeaderMap.Get("Location"))

	// Edit() as other, ADMIN PROTECTED
	res = as.HTML("/docs/%v/edit", doc.ID).Get()
	as.Equal(http.StatusFound, res.Code)
	as.Equal("/", res.HeaderMap.Get("Location"))

	// Update() as other, ADMIN PROTECTED
	res = as.HTML("/docs/%v", doc.ID).Put(doc)
	as.Equal(http.StatusFound, res.Code)
	as.Equal("/", res.HeaderMap.Get("Location"))
}
