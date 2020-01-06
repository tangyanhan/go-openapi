package openapi

import (
	"reflect"
	"strconv"
)

// Operation OperationObject
type Operation struct {
	method      string
	path        *Path
	Tags        []string     `json:"tags,omitempty"`
	Summary     string       `json:"summary,omitempty"`
	Description string       `json:"description,omitempty"`
	OperationID string       `json:"operationId,omitempty"`
	Parameters  []*Param     `json:"parameters,omitempty"`
	RequestBody *RequestBody `json:"requestBody,omitempty"`
	Responses   Responses    `json:"responses" validate:"required"`
	Deprecated  bool         `json:"deprecated,omitempty"`
}

// Metadata add metadata to operation
func (o *Operation) Metadata(operationID, summary, description string) *Operation {
	if operationID == "" && OperationGenerator != nil {
		operationID = OperationGenerator(o.method, o.path.path)
	}
	if operationID != "" {
		o.OperationID = operationID
	}

	o.Summary = summary
	o.Description = description
	return o
}

// Root returns document root for operation
func (o *Operation) Root() *OpenAPI {
	if o.path == nil {
		panic("no path is set for operation")
	}
	return o.path.Root()
}

// Returns with code
func (o *Operation) Returns(code int, description string, key string, v interface{}) *Operation {
	strCode := strconv.Itoa(code)
	if _, exists := o.Responses[strCode]; exists {
		panic("operation " + o.OperationID + " already returns code " + strCode)
	}
	r := o.newResponse(description, key, v)
	o.Responses[strCode] = r
	return o
}

// ReturnsNonJSON return something not json
func (o *Operation) ReturnsNonJSON(code int, description string,
	mimeType string, headers map[string]*Param, schema *Schema, example interface{}) *Operation {
	strCode := strconv.Itoa(code)
	if _, exists := o.Responses[strCode]; exists {
		panic("operation " + o.OperationID + " already returns code " + strCode)
	}
	o.Responses[strCode] = &Response{
		Description: description,
		Headers:     headers,
		Content: mediaTypeMap{
			mimeType: &MediaType{
				Schema:  schema,
				Example: example,
			},
		},
	}
	return o
}

// ReturnDefault add default response.
// A default response is the response to be used when none of defined codes match the situation.
func (o *Operation) ReturnDefault(description string, key string, v interface{}) *Operation {
	o.Responses["default"] = o.newResponse(description, key, v)
	return o
}

func (o *Operation) newResponse(description string, key string, v interface{}) *Response {
	schema := o.Root().MustGetSchema(key, v)
	return &Response{
		Description: description,
		Headers:     make(paramMap),
		Content: mediaTypeMap{
			MimeJSON: &MediaType{
				Schema:  schema,
				Example: v,
			},
		},
	}
}

// ReadJSON read object json from request body
func (o *Operation) ReadJSON(description string, required bool, key string, v interface{}) *Operation {
	schema := o.Root().MustGetSchema(key, v)
	o.RequestBody = &RequestBody{
		Description: description,
		Required:    required,
		Content: mediaTypeMap{
			MimeJSON: &MediaType{
				Schema:  schema,
				Example: v,
			},
		},
	}
	return o
}

// Read read raw body of any kind
func (o *Operation) Read(description string, required bool, mimeType string, example interface{}) *Operation {
	o.RequestBody = &RequestBody{
		Description: description,
		Required:    required,
		Content: mediaTypeMap{
			mimeType: &MediaType{
				Example: example,
			},
		},
	}
	return o
}

// AddParam with param in operation
func (o *Operation) AddParam(in ParamType, name, description string) *Param {
	if !in.IsValid() {
		panic("invalid param in " + in)
	}
	param := &Param{
		In:          in,
		Name:        name,
		Description: description,
	}
	// A path param is always required
	if in == PathParam {
		param.Required = true
		param.Schema = &Schema{
			Type: "string",
		}
	}
	o.Parameters = append(o.Parameters, param)
	return param
}

// WithParam add param to operation
func (o *Operation) WithParam(param *Param) *Operation {
	if !param.In.IsValid() {
		panic("invalid param in " + param.In)
	}
	o.Parameters = append(o.Parameters, param)
	return o
}

// WithPathParam add path param
func (o *Operation) WithPathParam(name, description string) *Operation {
	return o.WithParam(&Param{
		In:          PathParam,
		Name:        name,
		Description: description,
		Required:    true,
		Schema: &Schema{
			Type: "string",
		},
	})
}

// WithQueryParam add query param. Complex types of query param is not supported here(e.g., a struct or slice)
func (o *Operation) WithQueryParam(name, description string, example interface{}) *Operation {
	tv := reflect.TypeOf(example)
	typ, _ := kindToType(tv.Kind())
	return o.WithParam(&Param{
		In:          QueryParam,
		Name:        name,
		Description: description,
		Example:     example,
		Schema: &Schema{
			Type: typ,
		},
	})
}

// WithTags add tags
func (o *Operation) WithTags(tags ...string) *Operation {
	o.Tags = append(o.Tags, tags...)
	return o
}
