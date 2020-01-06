package openapi

import (
	"path"
	"reflect"
)

// Router is a extract of api path.
// router grouped in multiple levels can share parameters(path/query/header/etc)
type Router interface {
	Root() *OpenAPI
	WithParam(param *Param) Router
	WithPathParam(name, description string) Router
	WithTags(tags ...string) Router
	Route(path string, fn func(r Router)) Router
	// HTTP methods
	GET(path, summary, description string) *Operation
	PUT(path, summary, description string) *Operation
	POST(path, summary, description string) *Operation
	DELETE(path, summary, description string) *Operation
	HEAD(path, summary, description string) *Operation
	PATCH(path, summary, description string) *Operation
}

type router struct {
	root      *OpenAPI
	parent    *router
	path      string
	tags      []string
	params    []*Param
	paths     map[string]*Path
	subRoutes map[string]*router
}

// NewRouter create a new router
func NewRouter(root *OpenAPI) Router {
	return newRouter(root)
}

func newRouter(root *OpenAPI) *router {
	if root == nil {
		panic(ErrNoRoot)
	}
	return &router{
		root:      root,
		paths:     make(map[string]*Path),
		subRoutes: make(map[string]*router),
	}
}

// Root return document root
func (r *router) Root() *OpenAPI {
	return r.root
}

// GET short cut
func (r *router) GET(path, summary, description string) *Operation {
	return r.Method("get", path, summary, description)
}

// PUT put
func (r *router) PUT(path, summary, description string) *Operation {
	return r.Method("put", path, summary, description)
}

// POST post
func (r *router) POST(path, summary, description string) *Operation {
	return r.Method("post", path, summary, description)
}

func (r *router) DELETE(path, summary, description string) *Operation {
	return r.Method("delete", path, summary, description)
}

func (r *router) PATCH(path, summary, description string) *Operation {
	return r.Method("patch", path, summary, description)
}

func (r *router) HEAD(path, summary, description string) *Operation {
	return r.Method("head", path, summary, description)
}

// Method add method to router
func (r *router) Method(method, path, summary, description string) *Operation {
	apiPath, exists := r.paths[path]

	// Retrieve upstream to collect things back
	retriveUpstream := func(fn func(upstream *router)) {
		upstream := r
		for upstream != nil {
			fn(upstream)
			upstream = upstream.parent
		}
	}
	if !exists {
		newPath := &Path{
			root:       r.root,
			operations: make(opMap),
		}

		pathParts := []string{path}
		retriveUpstream(func(upstream *router) {
			pathParts = append(pathParts, upstream.path)
			newPath.Parameters = append(newPath.Parameters, upstream.params...)
		})

		reverse(pathParts)
		fullPath := joinPathParts(pathParts...)
		apiPath, exists = r.root.Paths[fullPath]
		if !exists {
			r.paths[path] = newPath
			r.root.Paths[fullPath] = newPath
			apiPath = newPath
		}
	}
	tags := make([]string, 0)
	retriveUpstream(func(upstream *router) {
		tags = append(tags, upstream.tags...)
	})
	op := apiPath.AddOperation(method)
	op.Summary = summary
	op.Description = description
	return op.WithTags(tags...)
}

// WithParam add param
func (r *router) WithParam(param *Param) Router {
	r.params = append(r.params, param)
	return r
}

// WithPathParam add path param
func (r *router) WithPathParam(name, description string) Router {
	return r.WithParam(&Param{
		Name:        name,
		In:          PathParam,
		Description: description,
		Required:    true,
		Schema: &Schema{
			Type: "string",
		},
	})
}

// WithTags add tag to the path
func (r *router) WithTags(tags ...string) Router {
	r.tags = append(r.tags, tags...)
	return r
}

// Route to sub paths. Remember that the returned router is newly created **sub** router
func (r *router) Route(path string, fn func(r Router)) Router {
	sub := newRouter(r.root)
	sub.parent = r
	sub.path = path
	r.subRoutes[path] = sub
	if fn != nil {
		fn(sub)
	}

	return r
}

func joinPathParts(parts ...string) string {
	return path.Join(parts...)
}

func reverse(slice interface{}) {
	rv := reflect.ValueOf(slice)
	swap := reflect.Swapper(slice)
	length := rv.Len()
	half := length / 2
	for i := 0; i < half; i++ {
		j := length - i - 1
		swap(i, j)
	}
}
