package openapi

import (
	"path"
	"reflect"
	"strings"
)

// DefaultOperationID provide default operation id generator
func DefaultOperationID(method, path string) string {
	method = strings.ToLower(method)
	pathParts := strings.Split(path, "/")
	var b strings.Builder
	b.WriteString(method)
	for _, part := range pathParts {
		p := strings.Trim(part, "{}")
		if p != part {
			b.WriteString("By")
		}
		b.WriteString(strings.Title(p))
	}
	return b.String()
}

func genInterfaceKey(v interface{}) string {
	tp := reflect.TypeOf(v)
	var prefix string
	for tp.PkgPath() == "" {
		switch tp.Kind() {
		case reflect.Ptr:
			tp = tp.Elem()
		case reflect.Slice:
			tp = tp.Elem()
			prefix += "array."
		}
	}

	// Full package path will confuse ref and make it unresolvable,
	// so we only keep package.TypeName as key
	fullPath := prefix + tp.PkgPath() + "." + tp.Name()
	key := prefix + path.Base(fullPath)
	return key
}
