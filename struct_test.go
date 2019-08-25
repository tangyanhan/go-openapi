package openapi

import (
	"testing"
)

// person
type person struct {
	Name   string   `json:"name" validate:"required,min=5,max=20"`
	Age    int      `json:"age" validate:"min=0,max=150"`
	Height float32  `json:"height" validate:"min=60,max=250.5"`
	Labels []string `json:"labels"`
}

func TestStructTags(t *testing.T) {
	s := person{
		"James", 16, 170.5, []string{"a", "b"},
	}
	schema, err := Interface(s)
	if err != nil {
		t.Fatal(err)
	}

	if len(schema.Required.Properties) != 1 || schema.Required.Properties[0] != "name" {
		t.Fatal(schema.Required.Properties)
	}
	nameProp, ok := schema.Properties["name"]
	if !ok {
		t.Fatal("name propertity not found")
	}
	if nameProp.Type != "string" {
		t.Fatalf("Expected name to be string, got %s", nameProp.Type)
	}

	if *nameProp.MinLength != 5 || *nameProp.MaxLength != 20 {
		t.Fatalf("MinLength=%d, MaxLength=%d", *nameProp.MinLength, *nameProp.MaxLength)
	}
	raw, err := json.Marshal(schema)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(raw))
}

func TestPtr(t *testing.T) {
	s := &person{
		"James", 16, 170.5, nil,
	}
	schema, err := Interface(s)
	if err != nil {
		t.Fatal(err)
	}

	if len(schema.Required.Properties) != 1 || schema.Required.Properties[0] != "name" {
		t.Fatal(schema.Required.Properties)
	}
	nameProp, ok := schema.Properties["name"]
	if !ok {
		t.Fatal("name propertity not found")
	}
	if nameProp.Type != "string" {
		t.Fatalf("Expected name to be string, got %s", nameProp.Type)
	}

	if *nameProp.MinLength != 5 || *nameProp.MaxLength != 20 {
		t.Fatalf("MinLength=%d, MaxLength=%d", *nameProp.MinLength, *nameProp.MaxLength)
	}
	raw, err := json.Marshal(schema)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(raw))
}

func TestStructSlice(t *testing.T) {
	s := [][]person{
		[]person{
			person{"James", 16, 170.5, nil},
			person{"Tom", 18, 174.0, nil},
		},
	}
	schema, err := Interface(s)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(schema)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(raw))
}

func TestStructPtrSlice(t *testing.T) {
	s := [][]*person{
		[]*person{
			&person{"James", 16, 170.5, nil},
			&person{"Tom", 18, 174.0, nil},
		},
	}
	schema, err := Interface(s)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(schema)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(raw))
}

func TestStructMap(t *testing.T) {
	s := map[string]person{
		"tom":   person{"Tom", 18, 174.0, nil},
		"james": person{"James", 16, 170.5, nil},
	}
	schema, err := Interface(s)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(schema)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(raw))
}

func TestEmptyStructMap(t *testing.T) {
	s := map[string]person{}
	schema, err := Interface(s)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(schema)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(raw))
}

func TestPtrMap(t *testing.T) {
	s := map[string]*person{
		"tom":   &person{"Tom", 18, 174.0, nil},
		"james": &person{"James", 16, 170.5, nil},
	}
	schema, err := Interface(s)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(schema)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(raw))
}

func TestEmptyMap(t *testing.T) {
	s := map[string]*person{}
	schema, err := Interface(s)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(schema)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(raw))
}

func TestPlainMap(t *testing.T) {
	s := map[string]int{}
	schema, err := Interface(s)
	if err != nil {
		t.Fatal(err)
	}
	if schema.AdditionalProperties.Type != "integer" {
		t.Fatal(schema.AdditionalProperties.Type)
	}
	if schema.AdditionalProperties.Format != "int64" {
		t.Fatal(schema.AdditionalProperties.Format)
	}
}

func TestEmbedStruct(t *testing.T) {
	a := struct {
		Name          string             `json:"name"`
		Age           int                `json:"age"`
		Labels        map[string]string  `json:"labels"`
		Family        map[string]*person `json:"family"`
		FamilyMembers []*person          `json:"familyMembers"`
	}{
		Name: "alice",
		Age:  16,
		Labels: map[string]string{
			"power":       "strong",
			"agile":       "medium",
			"inteligence": "weak",
		},
		Family: map[string]*person{
			"father": &person{
				Name: "zombie",
			},
		},
		FamilyMembers: []*person{
			&person{
				Name: "zombie",
			},
			&person{
				Name: "bob",
			},
		},
	}
	schema, err := Interface(&a)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(schema)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(raw))
}
