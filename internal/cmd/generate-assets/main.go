package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

var (
	outputTemplate = template.Must(template.New("output").Funcs(template.FuncMap{
		"printMultiLineString": printMultiLineString,
	}).Parse(`// Code generated by github.com/twpayne/chezmoi/internal/cmd/generate-assets. DO NOT EDIT.
{{- if .Tags}}
// +build {{ .Tags }}
{{- end }}

package cmd

func init() {
{{- range $key, $value := .Assets }}
	assets[{{ printf "%q" $key }}] = []byte("" +
		{{ printMultiLineString $value }})
{{- end }}
}`))

	output     = flag.String("o", "/dev/stdout", "output")
	trimPrefix = flag.String("trimprefix", "", "trim prefix")
	tags       = flag.String("tags", "", "tags")
)

func printMultiLineString(s []byte) string {
	sb := &strings.Builder{}
	for i, line := range bytes.Split(s, []byte{'\n'}) {
		if i != 0 {
			sb.WriteString(" +\n")
		}
		sb.WriteString(fmt.Sprintf("%q", append(line, '\n')))
	}
	return sb.String()
}

func run() error {
	flag.Parse()

	assets := make(map[string][]byte)
	for _, arg := range flag.Args() {
		var err error
		assets[strings.TrimPrefix(arg, *trimPrefix)], err = ioutil.ReadFile(arg)
		if err != nil {
			return err
		}
	}

	sb := &strings.Builder{}
	if err := outputTemplate.Execute(sb, struct {
		Tags   string
		Assets map[string][]byte
	}{
		Tags:   *tags,
		Assets: assets,
	}); err != nil {
		return err
	}

	formattedSource, err := format.Source([]byte(sb.String()))
	if err != nil {
		return err
	}

	return ioutil.WriteFile(*output, formattedSource, 0o666)
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}