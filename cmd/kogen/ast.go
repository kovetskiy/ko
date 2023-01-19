package main

import (
	"go/ast"
	"regexp"
	"strings"

	"github.com/iancoleman/strcase"
)

func astPath(field *ast.Field) string {
	if field.Tag == nil {
		return strcase.ToSnake(astFieldName(field))
	}
	allTags := astTags(field.Tag.Value)
	knownTags := []string{"yaml", "toml", "json"}
	for _, tag := range knownTags {
		value, ok := allTags[tag]
		if !ok || value == "" || value == "-" {
			continue
		}

		parts := strings.Split(value, ",")
		return parts[0]
	}

	return strcase.ToSnake(astFieldName(field))
}

func astFieldName(field *ast.Field) string {
	return field.Names[0].Name
}

func astTag(tags map[string]string, tag string, defaultValue string) string {
	if value, ok := tags[tag]; ok {
		return value
	}
	return defaultValue
}

func astTags(line string) map[string]string {
	line = strings.Trim(line, "`")

	chunks := regexp.MustCompile(`\s+`).Split(line, -1)

	values := make(map[string]string)
	for _, chunk := range chunks {
		keyValue := strings.SplitN(chunk, ":", 2)

		var key string
		var value string

		if len(keyValue) == 2 {
			key, value = keyValue[0], keyValue[1]
		} else {
			key = keyValue[0]
		}

		values[key] = strings.TrimSuffix(strings.TrimPrefix(value, `"`), `"`)
	}

	return values
}
