package openapi

import (
	"fmt"
	"log"
)

func ExampleNew() {
	o, err := New("3.0.0", sampleInfo)
	if err != nil {
		log.Fatal(err)
	}
	r := NewRouter(o)
	r.Route("/books", func(r Router) {
		r.GET("/", "List books", "List books").
			Returns(200, "Book content", "bookArray", []*Book{}).
			Returns(404, "Book not found", "replyError", &ReplyError{
				Code:    "book_not_found",
				Message: "The request book is not found",
			}).ReturnDefault("internal errors", "replyError", &ReplyError{
			Code:    "internal_error",
			Message: "an unknown error occurred in our end",
		})
		r.POST("/", "Add new book", "Add a new book").
			ReadJSON("JSON of book info", true, "book", &Book{}).
			Returns(200, "Book content", "book", &Book{}).
			Returns(404, "Book not found", "replyError", &ReplyError{
				Code:    "book_not_found",
				Message: "The request book is not found",
			})
		r.Route("/{id}", func(r Router) {
			r.WithPathParam("id", "ID of the book")
			r.GET("", "Get single book", "Info of a book").
				Returns(200, "Book content", "book", &Book{}).
				Returns(404, "Book not found", "replyError", &ReplyError{
					Code:    "book_not_found",
					Message: "The request book is not found",
				})
		})
	})

	raw, err := o.JSON()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(string(raw))
	// Output:
	// {"openapi":"3.0.0","info":{"title":"testing","version":"v1.0","termsOfService":"abcd","contact":{"name":"Ethan Tang","url":"example.com","email":"someone@example.com"},"license":{"name":"MIT License","url":"http://example.com/mit"}},"paths":{"/books":{"description":"","get":{"summary":"List books","description":"List books","responses":{"200":{"description":"Book content","content":{"application/json":{"schema":{"$ref":"#/components/schemas/bookArray"},"example":[]}}},"404":{"description":"Book not found","content":{"application/json":{"schema":{"$ref":"#/components/schemas/replyError"},"example":{"code":"book_not_found","message":"The request book is not found"}}}},"default":{"description":"internal errors","content":{"application/json":{"schema":{"$ref":"#/components/schemas/replyError"},"example":{"code":"internal_error","message":"an unknown error occurred in our end"}}}}}},"post":{"summary":"Add new book","description":"Add a new book","requestBody":{"description":"JSON of book info","content":{"application/json":{"schema":{"$ref":"#/components/schemas/book"},"example":{"name":"","author":"","date":""}}},"required":true},"responses":{"200":{"description":"Book content","content":{"application/json":{"schema":{"$ref":"#/components/schemas/book"},"example":{"name":"","author":"","date":""}}}},"404":{"description":"Book not found","content":{"application/json":{"schema":{"$ref":"#/components/schemas/replyError"},"example":{"code":"book_not_found","message":"The request book is not found"}}}}}},"summary":""},"/books/{id}":{"description":"","get":{"summary":"Get single book","description":"Info of a book","responses":{"200":{"description":"Book content","content":{"application/json":{"schema":{"$ref":"#/components/schemas/book"},"example":{"name":"","author":"","date":""}}}},"404":{"description":"Book not found","content":{"application/json":{"schema":{"$ref":"#/components/schemas/replyError"},"example":{"code":"book_not_found","message":"The request book is not found"}}}}}},"parameters":[{"name":"id","in":"path","description":"ID of the book","required":true,"schema":{"type":"string"}}],"summary":""}},"components":{"schemas":{"book":{"type":"object","properties":{"author":{"type":"string","maxLength":128,"minLength":1},"date":{"type":"string","format":"date"},"name":{"type":"string","maxLength":128,"minLength":1}},"required":["name","author"]},"bookArray":{"type":"array","items":{"type":"object","properties":{"author":{"type":"string","maxLength":128,"minLength":1},"date":{"type":"string","format":"date"},"name":{"type":"string","maxLength":128,"minLength":1}},"required":["name","author"]}},"replyError":{"type":"object","properties":{"code":{"type":"integer","description":"Error code of current string, read by program"},"message":{"type":"string","description":"Human friendly message that may help with the problem"}}}}}}
	//
}
