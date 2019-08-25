package openapi

import (
	"bytes"
	"strconv"
)

// SchemaRequired is a special representation for schema field "required",
// which can be either boolean or an array of propertity names
type SchemaRequired struct {
	Required   bool
	Properties []string
}

// MarshalJSON for special required field
func (s SchemaRequired) MarshalJSON() ([]byte, error) {
	if len(s.Properties) != 0 {
		var b bytes.Buffer
		b.WriteRune('[')
		for _, v := range s.Properties {
			if b.Len() != 1 {
				b.WriteRune(',')
			}
			b.WriteString(strconv.Quote(v))
		}
		b.WriteRune(']')
		return b.Bytes(), nil
	}
	if s.Required {
		return []byte("true"), nil
	}
	return []byte("false"), nil
}

// Schema SchemaObject
type Schema struct {
	root                 *OpenAPI
	key                  string
	Ref                  string             `json:"-"`
	Type                 string             `json:"type,omitempty"`
	Format               string             `json:"format,omitempty" validate:"oneof=int32 int64 float double byte binary date date-time password"`
	AllOf                []*Schema          `json:"allOf,omitempty"`
	OneOf                []*Schema          `json:"oneOf,omitempty"`
	AnyOf                []*Schema          `json:"anyOf,omitempty"`
	Not                  *Schema            `json:"not,omitempty"`
	Items                *Schema            `json:"items,omitempty"`
	Properties           map[string]*Schema `json:"properties,omitempty"`
	AdditionalProperties *Schema            `json:"additionalProperties,omitempty"`
	Description          string             `json:"description,omitempty"`
	Default              interface{}        `json:"default,omitempty"`
	Maximum              interface{}        `json:"maximum,omitempty"`
	Minimum              interface{}        `json:"minimum,omitempty"`
	MaxLength            *int64             `json:"maxLength,omitempty"`
	MinLength            *int64             `json:"minLength,omitempty"`
	Pattern              string             `json:"pattern,omitempty"`
	Required             *SchemaRequired    `json:"required,omitempty"`
	Enum                 []string           `json:"enum,omitempty"`
}

// SetRoot recursively set root for schema and all its related schemas
func (s *Schema) SetRoot(root *OpenAPI) {
	if s.root != nil {
		return
	}
	s.root = root
	for _, v := range s.AnyOf {
		v.SetRoot(root)
	}
	for _, v := range s.OneOf {
		v.SetRoot(root)
	}
	for _, v := range s.AnyOf {
		v.SetRoot(root)
	}
	if s.Not != nil {
		s.Not.SetRoot(root)
	}
	for _, v := range s.Properties {
		v.SetRoot(root)
	}
	if s.AdditionalProperties != nil {
		s.AdditionalProperties.SetRoot(root)
	}
}

// MarshalJSON turns
func (s Schema) MarshalJSON() ([]byte, error) {
	if s.Ref != "" {
		return []byte(`{"$ref":` + strconv.Quote(s.Ref) + `}`), nil
	}
	mirror := struct {
		Type                 string             `json:"type,omitempty"`
		Format               string             `json:"format,omitempty" validate:"oneof=int32 int64 float double byte binary date date-time password"`
		AllOf                []*Schema          `json:"allOf,omitempty"`
		OneOf                []*Schema          `json:"oneOf,omitempty"`
		AnyOf                []*Schema          `json:"anyOf,omitempty"`
		Not                  *Schema            `json:"not,omitempty"`
		Items                *Schema            `json:"items,omitempty"`
		Properties           map[string]*Schema `json:"properties,omitempty"`
		AdditionalProperties *Schema            `json:"additionalProperties,omitempty"`
		Description          string             `json:"description,omitempty"`
		Default              interface{}        `json:"default,omitempty"`
		Maximum              interface{}        `json:"maximum,omitempty"`
		Minimum              interface{}        `json:"minimum,omitempty"`
		MaxLength            *int64             `json:"maxLength,omitempty"`
		MinLength            *int64             `json:"minLength,omitempty"`
		Pattern              string             `json:"pattern,omitempty"`
		Required             *SchemaRequired    `json:"required,omitempty"`
		Enum                 []string           `json:"enum,omitempty"`
	}{
		Type:                 s.Type,
		Format:               s.Format,
		AllOf:                s.AllOf,
		OneOf:                s.OneOf,
		AnyOf:                s.AnyOf,
		Not:                  s.Not,
		Items:                s.Items,
		Properties:           s.Properties,
		AdditionalProperties: s.AdditionalProperties,
		Description:          s.Description,
		Default:              s.Default,
		Maximum:              s.Maximum,
		Minimum:              s.Minimum,
		MaxLength:            s.MaxLength,
		MinLength:            s.MinLength,
		Pattern:              s.Pattern,
		Required:             s.Required,
		Enum:                 s.Enum,
	}
	return json.Marshal(&mirror)
}

// NewSchema create new schema
func NewSchema(schemaType string) *Schema {
	return &Schema{
		Type:       schemaType,
		Properties: make(map[string]*Schema),
	}
}

// WithDescription add description
func (s *Schema) WithDescription(description string) *Schema {
	s.Description = description
	return s
}

// WithProperty add to a schema
func (s *Schema) WithProperty(name string, required bool, prop *Schema) *Schema {
	s.Properties[name] = prop
	if required {
		if s.Required == nil {
			s.Required = &SchemaRequired{}
		}
		s.Required.Properties = append(s.Required.Properties, name)
	}
	return s
}

// WithBasicProperty add a basic propertity
func (s *Schema) WithBasicProperty(name, propType, description string, required bool) *Schema {
	s.Properties[name] = &Schema{
		Type:        propType,
		Description: description,
	}
	if required {
		if s.Required == nil {
			s.Required = &SchemaRequired{}
		}
		s.Required.Properties = append(s.Required.Properties, name)
	}
	return s
}

// WithRequired make a schema to required: true or false
func (s *Schema) WithRequired(required bool) *Schema {
	if s.Required == nil {
		s.Required = &SchemaRequired{}
	}
	s.Required.Required = required
	return s
}

// WithOneOf set as one of.
func (s *Schema) WithOneOf(keyAndValues ...interface{}) *Schema {
	if len(keyAndValues)%2 != 0 {
		panic("invalid kv pair args")
	}

	n := len(keyAndValues) / 2
	for i := 0; i < n; i++ {
		k, v := keyAndValues[i*2], keyAndValues[i*2+1]
		key := k.(string)
		var (
			schema *Schema
			err    error
		)
		if s.root != nil {
			schema = s.root.MustGetSchema(key, v)
		} else {
			schema, err = Interface(v)
			if err != nil {
				panic(err)
			}
		}
		s.OneOf = append(s.OneOf, schema)
	}
	return s
}

// WithAnyOf set as any of
func (s *Schema) WithAnyOf(args ...interface{}) *Schema {
	if len(args) == 0 {
		panic(ErrNoOneOf)
	}
	for _, arg := range args {
		schema, err := Interface(arg)
		if err != nil {
			panic(err)
		}
		s.AnyOf = append(s.AnyOf, schema)
	}
	return s
}

// WithItems set array items
func (s *Schema) WithItems(schema *Schema) *Schema {
	s.Items = schema
	return s
}
