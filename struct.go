package openapi

// Provide utility to convert struct to OpenAPI document
import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Interface parse interface as schema
func Interface(v interface{}) (schema *Schema, err error) {
	// Try SchemaDoc provided by itself
	schemaProvider, ok := v.(SchemaDoc)
	if ok {
		schema = schemaProvider.SchemaDoc()
	} else {
		var err error
		schema, err = parseInterface(reflect.TypeOf(v), reflect.ValueOf(v))
		if err != nil {
			return nil, err
		}
	}
	return schema, nil
}

func parseMap(tv reflect.Type, rv reflect.Value) (*Schema, error) {
	schema := &Schema{
		Type:                 "object",
		AdditionalProperties: &Schema{},
	}
	elemType := tv.Elem()
	switch elemType.Kind() {
	case reflect.Struct, reflect.Ptr:
		var elemValue reflect.Value
		if len(rv.MapKeys()) != 0 {
			elemValue = rv.MapIndex(rv.MapKeys()[0])
		} else {
			elemValue = reflect.New(elemType).Elem()
		}
		s, err := parseInterface(elemType, elemValue)
		if err != nil {
			return nil, fmt.Errorf("failed to process map element %v", rv.Elem())
		}
		schema.AdditionalProperties = s
	default:
		typ, format := kindToType(elemType.Kind())
		schema.AdditionalProperties.Type = typ
		schema.AdditionalProperties.Format = format
	}

	return schema, nil
}

func kindToType(kind reflect.Kind) (typ string, format string) {
	switch kind {
	case reflect.Int32:
		return "integer", "int32"
	case reflect.Int, reflect.Int64:
		return "integer", "int64"
	case reflect.Float32, reflect.Float64:
		return "number", ""
	case reflect.String:
		return "string", ""
	case reflect.Bool:
		return "boolean", ""
	default:
		return "object", ""
	}
}

func parseInterface(tv reflect.Type, rv reflect.Value) (schema *Schema, err error) {
	method := rv.MethodByName("SchemaDoc")
	if method.IsValid() {
		values := method.Call(nil)
		if len(values) != 1 {
			return nil, ErrInvalidSchemaDoc
		}
		v := values[0].Interface()
		s, ok := v.(*Schema)
		if !ok {
			return nil, ErrInvalidSchemaDoc
		}
		return s, nil
	}
	switch tv.Kind() {
	case reflect.Struct:
		return parseStruct(tv, rv)
	case reflect.Slice:
		elemType := tv.Elem()
		var elemValue reflect.Value
		if !rv.IsNil() && rv.Len() != 0 {
			elemValue = rv.Index(0)
		} else {
			elemValue = reflect.New(tv.Elem()).Elem()
		}
		schema, err := parseInterface(elemType, elemValue)
		if err != nil {
			return nil, err
		}
		return &Schema{
			Type:  "array",
			Items: schema,
		}, nil
	case reflect.Map:
		return parseMap(tv, rv)
	case reflect.Ptr:
		for tv.Kind() == reflect.Ptr {
			tv = tv.Elem()
			if !rv.IsNil() {
				rv = rv.Elem()
			} else {
				rv = reflect.New(tv).Elem()
			}
		}
		return parseInterface(tv, rv)
	default:
		typ, format := kindToType(tv.Kind())
		return &Schema{
			Type:   typ,
			Format: format,
		}, nil
	}
}

func parseStruct(tv reflect.Type, rv reflect.Value) (schema *Schema, err error) {
	schema = &Schema{
		Type:       "object",
		Properties: make(map[string]*Schema),
	}
	for i := 0; i < tv.NumField(); i++ {
		f := tv.Field(i)
		var v reflect.Value
		if rv.Kind() == reflect.Ptr {
			v = reflect.New(f.Type).Elem()
		} else {
			v = rv.Field(i)
		}
		jsonTag, ok := f.Tag.Lookup("json")
		if !ok {
			continue
		}
		if jsonTag == "-" {
			continue
		}
		s, err := parseInterface(f.Type, v)
		if err != nil {
			return nil, fmt.Errorf("error parsing field %s", f.Name)
		}
		// Embeded tag
		if jsonTag == "," || jsonTag == ",inline" {
			schema.AllOf = append(schema.AllOf, s)
			continue
		}
		p := strings.Split(jsonTag, ",")
		propName := p[0]

		// Parse custom tags
		if err := parseCustomTags(f.Tag, s); err != nil {
			return nil, err
		}
		// Required is special
		vTag, ok := f.Tag.Lookup("validate")
		var required bool
		if ok {
			required, err = parseValidateTag(vTag, s)
			if err != nil {
				return nil, err
			}
		}
		schema.WithProperty(propName, required, s)
	}
	return schema, nil
}

func parseCustomTags(tag reflect.StructTag, schema *Schema) error {
	tagParsers := map[string]func(v string, schema *Schema) error{
		"description": func(v string, schema *Schema) error {
			schema.Description = v
			return nil
		},
		"format": func(v string, schema *Schema) error {
			schema.Format = v
			return nil
		},
		"pattern": func(v string, schema *Schema) error {
			schema.Pattern = v
			return nil
		},
		"enum": func(v string, schema *Schema) error {
			enums := strings.Split(v, "|")
			if len(enums) == 0 {
				return errors.New("no enum values")
			}
			schema.Enum = enums
			return nil
		},
		"default": func(v string, schema *Schema) error {
			defaultValue, err := tagToValue(schema.Type, v)
			if err != nil {
				return fmt.Errorf("error with default value:%s", err.Error())
			}
			schema.Default = defaultValue
			return nil
		},
	}
	for key, fn := range tagParsers {
		v, ok := tag.Lookup(key)
		if !ok {
			continue
		}
		if err := fn(v, schema); err != nil {
			return err
		}
	}
	return nil
}

// convert string tag value to corresponding value
func tagToValue(schemaType string, v string) (interface{}, error) {
	switch schemaType {
	case "integer":
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse value %s as %s:%s", v, schemaType, err.Error())
		}
		return i, nil
	case "number":
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse value %s as %s:%s", v, schemaType, err.Error())
		}
		return f, nil
	case "boolean":
		switch v {
		case "true":
			return true, nil
		case "false":
			return false, nil
		default:
			return nil, fmt.Errorf("failed to parse default value %s as %s", v, schemaType)
		}
	case "string":
		return v, nil
	default:
		return nil, fmt.Errorf("unknown default value %s for type %s", v, schemaType)
	}
}

// parse tags from golang validator
func parseValidateTag(vTag string, schema *Schema) (required bool, err error) {
	if vTag == "" || vTag == "-" {
		return false, nil
	}
	parts := strings.Split(vTag, ",")
	for _, p := range parts {
		if p == "required" {
			required = true
			continue
		}

		var v string
		var isMax bool
		if strings.HasPrefix(p, "max=") {
			isMax = true
			v = strings.TrimPrefix(p, "max=")
		} else if strings.HasPrefix(p, "min=") {
			v = strings.TrimPrefix(p, "min=")
		} else {
			continue
		}

		switch schema.Type {
		case "string":
			value, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return false, fmt.Errorf("failed to parse tag of value %s:%s", v, err.Error())
			}

			if isMax {
				schema.MaxLength = &value
			} else {
				schema.MinLength = &value
			}

		case "integer":
			value, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return false, fmt.Errorf("failed to parse tag of value %s:%s", v, err.Error())
			}

			if isMax {
				schema.Maximum = value
			} else {
				schema.Minimum = value
			}

		case "number":
			value, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return false, fmt.Errorf("failed to parse max tag of value %s:%s", v, err.Error())
			}

			if isMax {
				schema.Maximum = value
			} else {
				schema.Minimum = value
			}

		default:
			return false, fmt.Errorf("unknown max value %s for type %s", v, schema.Type)
		}
	}
	return required, nil
}
