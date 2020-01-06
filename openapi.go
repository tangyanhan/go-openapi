/*Package openapi provide OpenAPI 3.0 support for Go
In this package, there are many util functions that looks chained, but some are not.
If a method starts with "Add", it will return the instance that has just been added.
Otherwise, methods start with "With" returns the instance itself.
*/
package openapi

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/go-playground/validator"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// SchemaDoc provide schema for a struct
type SchemaDoc interface {
	SchemaDoc() *Schema
}

// APIExample provider
type APIExample interface {
	Example()
}

// OperationGenerator provide operationId when no operation id given
var OperationGenerator = DefaultOperationID

// Errors that may be returned or raised
var (
	ErrNoRoot           = errors.New("root not initialized with NewRoot")
	ErrNoOneOf          = errors.New("oneOf has no alternatives")
	ErrInvalidSchemaDoc = errors.New("invalid SchemaDoc method given")
)

// Supported mime types when using shortcuts
const (
	MimeJSON = "application/json"
	MimeYAML = "application/yaml"
)

// New create OpenAPI document object and set it as global document root
func New(version string, info Info) (*OpenAPI, error) {
	if !strings.HasPrefix(version, "3.") {
		return nil, fmt.Errorf("only openapi 3.x is supported")
	}
	err := validator.New().Struct(info)
	if err != nil {
		return nil, err
	}
	components := &Components{
		Schemas:       make(schemaMap),
		Responses:     make(respMap),
		Parameters:    make(paramMap),
		Examples:      make(exampleMap),
		RequestBodies: make(reqBodyMap),
		Headers:       make(paramMap),
	}

	return &OpenAPI{
		OpenAPI:    version,
		Info:       info,
		Paths:      make(pathMap),
		Components: components,
	}, nil
}

// YAML convert document to yaml.
// We only have json tags in structs, so special lib must be used
func (o *OpenAPI) YAML() ([]byte, error) {
	return yaml.Marshal(o)
}

// JSON marshal document as JSON
func (o OpenAPI) JSON() ([]byte, error) {
	return json.Marshal(o)
}

// GetSchema return schema ref if exists
func (o *OpenAPI) GetSchema(key string) *Schema {
	_, ok := o.Components.Schemas[key]
	if ok {
		return &Schema{
			key: key,
			Ref: "#/components/schemas/" + key,
		}
	}
	return nil
}

// MustGetSchema create schema for struct when necessary and always return a ref format
func (o *OpenAPI) MustGetSchema(key string, v interface{}) *Schema {
	if v == nil {
		return &Schema{
			root: o,
		}
	}
	schema := o.GetSchema(key)
	if schema == nil {
		var err error
		schema, err = Interface(v)
		if err != nil {
			panic(err)
		}
		if key == "" {
			key = genInterfaceKey(v)
		}
		schema = o.AddSchema(key, schema)
	}
	return schema
}

// AddSchema add schema to global components, and return a ref
func (o *OpenAPI) AddSchema(key string, schema *Schema) *Schema {
	schema.root = o
	o.Components.Schemas[key] = schema
	return &Schema{
		key: key,
		Ref: "#/components/schemas/" + key,
	}
}

// AddParam add param definition to global components
func (o *OpenAPI) AddParam(key string, param *Param) string {
	param.root = o
	o.Components.Parameters[key] = param
	return "#/components/parameters/" + key
}

// GetParam return param
func (o *OpenAPI) GetParam(key string) *Param {
	param, ok := o.Components.Parameters[key]
	if !ok {
		panic("failed to find param with key:" + key)
	}
	return param
}

// AddHeader add param definition to global components
func (o *OpenAPI) AddHeader(key string, param *Param) string {
	param.root = o
	param.In = "header"
	o.Components.Headers[key] = param
	return "#/components/headers/" + key
}

// GetHeader return param
func (o *OpenAPI) GetHeader(key string) *Param {
	param, ok := o.Components.Headers[key]
	if !ok {
		panic("failed to find param with key:" + key)
	}
	return param
}

// AddPath to OpenAPI paths section
func (o *OpenAPI) AddPath(path, summary, description string) *Path {
	if _, exists := o.Paths[path]; exists {
		panic("path already exists:" + path)
	}
	p := &Path{
		root:        o,
		path:        path,
		Summary:     summary,
		Description: description,
		operations:  make(opMap),
	}
	o.Paths[path] = p
	return p
}

// AddOperation add operation to path
func (p *Path) AddOperation(method string) *Operation {
	method = strings.ToLower(method)
	if _, exists := p.operations[method]; exists {
		panic("operation for path " + p.path + " already exists method " + method)
	}
	op := &Operation{
		method:    method,
		path:      p,
		Responses: make(Responses),
	}
	p.operations[method] = op
	return op
}

// NewPathParam create new path param
func NewPathParam(name, description string) *Param {
	return &Param{
		Name:        name,
		In:          PathParam,
		Description: description,
	}
}

// NewQueryParam create new query param
func NewQueryParam(name, description string, example interface{}) *Param {
	tv := reflect.TypeOf(example)
	typ, _ := kindToType(tv.Kind())
	return &Param{
		In:          QueryParam,
		Name:        name,
		Description: description,
		Example:     example,
		Schema: &Schema{
			Type: typ,
		},
	}
}

// SetRequired make param as required
func (p *Param) SetRequired() *Param {
	p.Required = true
	return p
}

// AllowEmpty make param allow empty
func (p *Param) AllowEmpty() *Param {
	p.AllowEmptyValue = true
	return p
}

// SetDeprecated make param as deprecated
func (p *Param) SetDeprecated() *Param {
	p.Deprecated = true
	return p
}

// WithStruct add struct schema for param
func (p *Param) WithStruct(v interface{}) *Param {
	schema, err := Interface(v)
	if err != nil {
		panic(err)
	}
	p.Schema = schema
	return p
}

// WithSchema add schema
func (p *Param) WithSchema(s *Schema) *Param {
	p.Schema = s
	return p
}
