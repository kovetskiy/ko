package main

import (
	_ "embed"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"text/template"

	"github.com/docopt/docopt-go"
)

var (
	version = "[manual build]"
	usage   = "kogen " + version + `

Kogen generates documentation for configuration files based on github.com/kovetskiy/ko.

Usage:
  kogen [options] <dir> <struct>
  kogen -h | --help
  kogen --version

Options:
  -t --template <file>  Use specified template file instead of the default one.
  -j --json             Output JSON instead of Markdown.
  <dir>                 Directory with Go source code.
  <struct>              Name of structure to generate documentation for.
  -h --help             Show this screen.
  --version             Show version.
`
)

var (
	//go:embed doc.template
	templateDocumentation string
)

func main() {
	args, err := docopt.Parse(usage, nil, true, version, false)
	if err != nil {
		panic(err)
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(
		fset,
		args["<dir>"].(string),
		nil,
		parser.ParseComments,
	)
	if err != nil {
		panic(err)
	}

	generator := &Generator{
		structs: make(map[string]*Struct),
	}
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			ast.Walk(generator, file)
		}
	}

	target, ok := generator.structs[args["<struct>"].(string)]
	if !ok {
		log.Fatalln("type not found:", args["<struct>"].(string))
	}

	fields := generator.generate(target)

	if args["--json"].(bool) {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		err = encoder.Encode(fields)
		if err != nil {
			log.Fatalln(err)
		}

		return
	}

	tpl := template.New("")
	tpl = tpl.Funcs(template.FuncMap{
		"backtick": func(s string) string {
			if s == "" {
				return "`<no value>`"
			}

			return "`" + s + "`"
		},
	})
	tpl = template.Must(tpl.Parse(templateDocumentation))
	err = tpl.Execute(os.Stdout, map[string][]StructField{"Fields": fields})
	if err != nil {
		log.Fatalln(err)
	}
}
