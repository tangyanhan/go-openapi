/*Package openapi provide OpenAPI 3.0 support for Go*/
package openapi

type pathMap map[string]*Path
type opMap map[string]*Operation

type respMap map[string]*Response

// OpenAPI document structure
type OpenAPI struct {
	OpenAPI    string      `json:"openapi"`
	Info       Info        `json:"info"`
	Servers    []Server    `json:"servers,omitempty"`
	Paths      pathMap     `json:"paths"`
	Components *Components `json:"components,omitempty"`
}

// Info of global document
type Info struct {
	Title          string   `json:"title" validate:"required"`
	Version        string   `json:"version" validate:"required"`
	TermsOfService string   `json:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty"`
	License        *License `json:"license,omitempty"`
}

// Server server object
type Server struct {
	URL         string                    `json:"url" validate:"required"`
	Description string                    `json:"description,omitempty"`
	Variables   map[string]ServerVariable `json:"variables,omitempty"`
}

// ServerVariable is used to replace some things in url schema
type ServerVariable struct {
	Enum        []string `json:"enum,omitempty"`
	Description string   `json:"description,omitempty"`
	Default     string   `json:"default" validate:"required"`
}

// Contact info
type Contact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// License info
type License struct {
	Name string `json:"name" validate:"required"`
	URL  string `json:"url,omitempty"`
}

// Path definition
type Path struct {
	root        *OpenAPI
	path        string
	Summary     string `json:"summary"`
	Description string `json:"description"`
	operations  opMap
	Parameters  []*Param `json:"parameters,omitempty"`
}

// MarshalJSON marshal path
func (p Path) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"summary":     p.Summary,
		"description": p.Description,
	}
	if len(p.Parameters) != 0 {
		m["parameters"] = p.Parameters
	}
	for method, op := range p.operations {
		m[method] = op
	}
	return json.Marshal(m)
}

// Root document of whole application
func (p *Path) Root() *OpenAPI {
	if p.root == nil {
		panic("no root for path " + p.path)
	}
	return p.root
}

// Param ParameterObject
type Param struct {
	root *OpenAPI
	// Fixed fields
	Name            string    `json:"name" validate:"required"`
	In              ParamType `json:"in" validate:"required,oneof=query header path cookie"`
	Description     string    `json:"description,omitempty"`
	Required        bool      `json:"required"`
	Deprecated      bool      `json:"deprecated,omitempty"`
	AllowEmptyValue bool      `json:"allowEmptyValue,omitempty"`
	// Below are optional fields
	Schema   *Schema             `json:"schema,omitempty"`
	Example  interface{}         `json:"example,omitempty"`
	Examples map[string]*Example `json:"examples,omitempty"`
}

// Example ExampleObj
type Example struct {
	Summary     string      `json:"summary,omitempty"`
	Description string      `json:"description,omitempty"`
	Value       interface{} `json:"value"`
}

// RequestBody request body object
type RequestBody struct {
	Description string `json:"description,omitempty"`
	// MIME-Type -> MediaTypeObject
	Content  mediaTypeMap `json:"content" validate:"required"`
	Required bool         `json:"required,omitempty"`
}

// MediaType media type object
type MediaType struct {
	Schema   *Schema             `json:"schema,omitempty"`
	Example  interface{}         `json:"example,omitempty"`
	Examples map[string]*Example `json:"examples,omitempty"`
}

// Responses is actually a map
type Responses map[string]*Response
type paramMap map[string]*Param
type mediaTypeMap map[string]*MediaType

// Response response object
type Response struct {
	Description string       `json:"description"`
	Headers     paramMap     `json:"headers,omitempty"`
	Content     mediaTypeMap `json:"content,omitempty"`
}

type schemaMap map[string]*Schema
type exampleMap map[string]*Example
type reqBodyMap map[string]*RequestBody

// Components object
type Components struct {
	Schemas       schemaMap        `json:"schemas,omitempty"`
	Responses     respMap          `json:"responses,omitempty"`
	Parameters    paramMap         `json:"parameters,omitempty"`
	Examples      exampleMap       `json:"examples,omitempty"`
	RequestBodies reqBodyMap       `json:"requestBodies,omitempty"`
	Headers       paramMap         `json:"-"`
	Links         map[string]*Link `json:"links,omitempty"`
}

// Link to a resuable object
type Link struct {
	OperationRef string            `json:"operationRef,omitempty"`
	OperationID  string            `json:"operationId,omitempty"`
	Parameters   map[string]*Param `json:"parameters,omitempty"`
	RequestBody  interface{}       `json:"requestBody,omitempty"`
	Description  string            `json:"description,omitempty"`
}

// ParamType param "in" request
type ParamType string

// Valid param types
const (
	PathParam   ParamType = "path"
	QueryParam  ParamType = "query"
	HeaderParam ParamType = "header"
	CookieParam ParamType = "cookie"
)

// IsValid param
func (p ParamType) IsValid() bool {
	switch p {
	case PathParam, QueryParam, HeaderParam, CookieParam:
		return true
	default:
		return false
	}
}
