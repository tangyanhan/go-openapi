# go-openapi
Shabby yet useful OpenAPI 3.0 support for Go

# Install

```
go get -u github.com/tangyanhan/go-openapi
```

# Quick Start

Organize your API routes like a tree:
```
	o, err := New("3.0.0", sampleInfo)
	if err != nil {
		log.Fatal(err)
	}
	r := NewRouter(o)
	r.Route("/{namespace}/books", func(r Router) {
    r.WithPathParam("namespace", "Namespace of the bookstore")
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
	})
  apiYAML, _ := o.YAML()
```
You don't have to write all arguments again and again, simply put common parameters in upstream branch.

The schemas of data to be read/written will be collected into API document automatically.

# Interface Schema

The schema of a given interface can be generated either automatically or via implementating ```SchemaDoc``` interface.

### Automatic Schema For Interface

Data can be provided either in struct or pointer. The package will collected all it's info and generate a schema.

During this process,  field tags will be used to generate propertity schemas::

```json``` tags will be used as the propertity name.

```doc``` tags will be used to generate these values for schema: ```enum, description, format and pattern```.
```go
   {
     kind string `json:"kind" doc="enum=a|b|c,description=blablabla"`
     age int `json:"age" doc="pattern=^abc$"`
   }
```

```validate``` tags will be used to generate values for schema: ```minimum, maximum, min, max, required```. If ```required,min,max``` is set, corresponding values will be set.

### Native Schema Generation

A type may implement ```openapi.SchemaDoc``` to provide a schema manually. If the implementation is found, the schema it returns will be used, instead of automatic schema generation.

```go
// ReplyError error reply
type ReplyError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (r ReplyError) SchemaDoc() *Schema {
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
```

# Known Issues

* The final document is not likely to be in common order.
* The schema for a interface will always be put into ```#/components/schemas```
* Parameters/Responses are not reused with ```ref``` in the document generated automatically
* If a type contain nested type, such like a map with struct as values, the struct schema is not likely to be in the ```ref``` style, it will be nested in the definition

This package has not been fully tested and covered. I will keep on updating it on my own needs in production.