package main

import (
	"io"
	"text/template"
)

var tmpl = template.Must(template.New("setter").Parse(
	`package {{.Package}} 
{{ range $message := .Messages }} {{ range $field := $message.Fields }}
func (t *{{ $message.Name }}) Set{{ $field.Name }}({{ $field.VarName }} {{ $field.VarType }}) {
    t.{{ $field.Name }} = {{ $field.VarName }}
}
{{ end }}{{ end }}`))

type setterFile struct {
	Package  string
	Name     string
	All      bool
	Messages []*setterMessage
}

func (file *setterFile) into(into io.Writer) error {
	return tmpl.Execute(into, file)
}

type setterMessage struct {
	Name   string
	All    bool
	Fields []*setterField
}

type setterField struct {
	Name    string
	VarName string
	VarType string
}
