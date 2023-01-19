package main

import (
	"go/ast"
	"go/token"
	"strings"
)

type StructField struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Path         string `json:"path"`
	DefaultValue string `json:"default_value"`
	Env          string `json:"env"`
	Required     string `json:"required"`
}

type Struct struct {
	Name   string
	Fields []*ast.Field
}

type Generator struct {
	structs map[string]*Struct
}

func (generator *Generator) generate(target *Struct, stack ...string) []StructField {
	result := []StructField{}

	for _, field := range target.Fields {
		switch fieldType := field.Type.(type) {
		case *ast.Ident:
			if _, ok := generator.structs[fieldType.Name]; ok {
				result = append(
					result,
					generator.generate(
						generator.structs[fieldType.Name],
						push(stack, astPath(field))...,
					)...,
				)
			} else {
				result = append(
					result,
					generator.getField(field, fieldType, stack...),
				)
			}

		case *ast.StructType:
			result = append(
				result,
				generator.generate(
					&Struct{
						Fields: fieldType.Fields.List,
					},
					push(stack, astPath(field))...,
				)...,
			)
		}
	}

	return result
}

func (generator *Generator) getField(
	field *ast.Field,
	fieldType *ast.Ident,
	stack ...string,
) StructField {
	fieldName := astFieldName(field)
	tags := astTags(field.Tag.Value)

	return StructField{
		Name:         fieldName,
		Type:         fieldType.Name,
		Path:         strings.Join(push(stack, astPath(field)), "."),
		DefaultValue: astTag(tags, "default", ""),
		Required:     astTag(tags, "required", "false"),
		Env:          astTag(tags, "env", ""),
	}
}

func (generator *Generator) Visit(node ast.Node) ast.Visitor {
	switch node := node.(type) {
	case *ast.Package:
		return generator
	case *ast.File:
		return generator
	case *ast.GenDecl:
		if node.Tok == token.TYPE {
			return generator
		}
	case *ast.TypeSpec:
		switch _type := node.Type.(type) {
		case *ast.StructType:
			generator.structs[node.Name.Name] = &Struct{
				Name:   node.Name.Name,
				Fields: _type.Fields.List,
			}
		}

		return generator
	}
	return nil
}
