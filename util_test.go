package openapi

import (
	"testing"
)

func TestGenOperationID(t *testing.T) {
	testCases := []struct {
		desc   string
		method string
		in     string
		expect string
	}{
		{
			desc:   "normal",
			method: "get",
			in:     "/books/",
			expect: "getBooks",
		},
		{
			desc:   "with-param",
			method: "get",
			in:     "/books/{id}",
			expect: "getBooksById",
		},
		{
			desc:   "with-multiple-param",
			method: "get",
			in:     "/a/{a-id}/b/{b-id}",
			expect: "getAByA-IdBByB-Id",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			out := DefaultOperationID(tC.method, tC.in)
			if out != tC.expect {
				t.Fatal("Expect=", tC.expect, "Got=", out)
			}
		})
	}
}

func Test_genInterfaceKey(t *testing.T) {
	tests := []struct {
		name    string
		v       interface{}
		want    string
		wantErr bool
	}{
		{
			name: "Interface",
			v:    &person{},
			want: "go-openapi.person",
		},
		{
			name: "Struct",
			v:    person{},
			want: "go-openapi.person",
		},
		{
			name: "Slice",
			v:    []person{},
			want: "array.go-openapi.person",
		},
		{
			name: "SlicePtr",
			v:    []*person{},
			want: "array.go-openapi.person",
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := genInterfaceKey(tt.v)
			if got != tt.want {
				t.Errorf("genInterfaceKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
