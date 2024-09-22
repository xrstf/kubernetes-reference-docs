/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package generators

import (
	"bytes"
	"html/template"
	"io"

	"github.com/Masterminds/sprig/v3"
)

var templates *template.Template

func init() {
	var err error

	templates, err = template.
		New("base").
		Funcs(sprig.FuncMap()).
		Funcs(template.FuncMap{}).
		ParseGlob("generators/templates/*")
	if err != nil {
		panic(err)
	}
}

func renderTemplateTo(dst io.Writer, filename string, data any) error {
	return templates.ExecuteTemplate(dst, filename, data)
}

func renderTemplate(filename string, data any) (template.HTML, error) {
	var buf bytes.Buffer

	err := renderTemplateTo(&buf, filename, data)
	if err != nil {
		return "", err
	}

	return template.HTML(buf.String()), nil
}
