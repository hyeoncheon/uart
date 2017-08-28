package models_test

// Test coverage: 100% (without interface methods)

import (
	"github.com/hyeoncheon/uart/models"
)

var doc = &models.Doc{
	Type:        "Manual",
	Category:    "Category1",
	Subject:     "Subject1",
	Slug:        "title-of-the-document",
	Title:       "Title of the Document",
	Content:     "Content of the document",
	IsPublished: true,
}

func (ms *ModelSuite) Test_Doc() {
	verrs, err := models.DB.ValidateAndCreate(member)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	doc.MemberID = member.ID
	verrs, err = models.DB.ValidateAndCreate(doc)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	ms.Equal(doc.Title, doc.String())
	ms.Equal("Dummy Member", doc.AuthorName())
}
