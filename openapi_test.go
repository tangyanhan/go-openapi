package openapi

import (
	"testing"
)

var sampleInfo = Info{
	Title:          "testing",
	Version:        "v1.0",
	TermsOfService: "abcd",
	License: &License{
		"MIT License",
		"http://example.com/mit",
	},
	Contact: &Contact{
		Name:  "Ethan Tang",
		URL:   "example.com",
		Email: "someone@example.com",
	},
}

func TestNew(t *testing.T) {
	o, err := New("3.0.0", sampleInfo)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	expect := `{"openapi":"3.0.0","info":{"title":"testing","version":"v1.0","termsOfService":"abcd","contact":{"name":"Ethan Tang","url":"example.com","email":"someone@example.com"},"license":{"name":"MIT License","url":"http://example.com/mit"}},"paths":{},"components":{}}`
	if expect != string(raw) {
		t.Fatal("Got:\n", string(raw))
	}
}

func TestPathOp(t *testing.T) {
	o, err := New("3.0.0", sampleInfo)
	if err != nil {
		t.Fatal(err)
	}
	p := o.AddPath("/books", "Books operation", "操作书籍")
	p.AddOperation("get").Metadata("getBooks", "List books", "List books with info")
	p.AddOperation("put").Metadata("putBook", "Create books", "Put a single book to books")
	raw, err := json.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(raw))
}

// Book info
type Book struct {
	Name   string `json:"name" validate:"required,min=1,max=128"`
	Author string `json:"author" validate:"required,min=1,max=128"`
	Date   string `json:"date" doc:"format=date"`
}

// ReplyError error reply
type ReplyError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (r *ReplyError) SchemaDoc() *Schema {
	return &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"code": &Schema{
				Type:        "integer",
				Description: "Error code of current string, read by program",
			},
			"message": &Schema{
				Type:        "string",
				Description: "Human friendly message that may help with the problem",
			},
		},
	}
}

func TestParamSchema(t *testing.T) {
	o, err := New("3.0.0", sampleInfo)
	if err != nil {
		t.Fatal(err)
	}
	path := o.AddPath("/books/{id}", "Books Operation", "Operate on books")
	op := path.AddOperation("get").Metadata("", "Get Single Book", "Get full info of a book")
	op.AddParam(PathParam, "id", "ID of the book")
	op.Returns(200, "Book content", "book", &Book{}).Returns(404, "Book not found", "replyError", &ReplyError{
		Code:    "book_not_found",
		Message: "The request book is not found",
	})
	path = o.AddPath("/books/", "Books Operation", "Operation on books")
	op = path.AddOperation("post").Metadata("", "Add a new book", "Add a new book to the store")
	op.ReadJSON("JSON of book info", true, "book", &Book{}).Returns(200, "Book content", "book", &Book{}).Returns(404, "Book not found", "replyError", &ReplyError{
		Code:    "book_not_found",
		Message: "The request book is not found",
	})
	raw, err := json.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(raw))
}
