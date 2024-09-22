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
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kubernetes-sigs/reference-docs/gen-apidocs/generators/api"
)

type TOCItem struct {
	Level       int
	Title       template.HTML
	Link        string
	File        string
	SubSections []*TOCItem
}

func (ti *TOCItem) ToHTML() template.HTML {
	rendered, err := renderTemplate("toc-item.html", ti)
	if err != nil {
		panic(err)
	}

	return rendered
}

type TOC struct {
	Title    string
	Sections []*TOCItem
}

type HTMLWriter struct {
	Config *api.Config
	TOC    TOC

	// currentTOCItem is used to remember the current item between
	// calls to e.g. WriteResourceCategory() followed by WriteResource().
	currentTOCItem *TOCItem
}

func NewHTMLWriter(config *api.Config, title string) (DocWriter, error) {
	writer := HTMLWriter{
		Config: config,
		TOC: TOC{
			Title:    title,
			Sections: []*TOCItem{},
		},
	}

	return &writer, nil
}

func (h *HTMLWriter) WriteOverview() error {
	filename := "_overview.html"
	if err := writeStaticFile(filename, h.sectionHeading("API Overview")); err != nil {
		return err
	}

	item := TOCItem{
		Level: 1,
		Title: "Overview",
		Link:  "api-overview",
		File:  filename,
	}
	h.TOC.Sections = append(h.TOC.Sections, &item)
	h.currentTOCItem = &item

	return nil
}

func (h *HTMLWriter) WriteAPIGroupVersions(gvs api.GroupVersions) error {
	groups := api.ApiGroups{}
	for group := range gvs {
		groups = append(groups, api.ApiGroup(group))
	}
	sort.Sort(groups)

	tplGroups := []map[string]any{}

	for _, group := range groups {
		versionList := gvs[group.String()]
		sort.Sort(versionList)
		var versions []string
		for _, v := range versionList {
			versions = append(versions, v.String())
		}

		tplGroups = append(tplGroups, map[string]any{
			"group":    group,
			"versions": versions,
		})
	}

	fn := "_api_groups.html"
	path := filepath.Join(api.IncludesDir, fn)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := renderTemplateTo(f, "api-groups.html", map[string]any{
		"groups": tplGroups,
	}); err != nil {
		return err
	}

	item := TOCItem{
		Level: 1,
		Title: "API Groups",
		Link:  "api-groups",
		File:  fn,
	}
	h.TOC.Sections = append(h.TOC.Sections, &item)
	h.currentTOCItem = &item

	return nil
}

func (h *HTMLWriter) WriteResourceCategory(name, file string) error {
	if err := writeStaticFile("_"+file+".html", h.resourceCategoryHeading(name)); err != nil {
		return err
	}

	link := strings.ReplaceAll(strings.ToLower(name), " ", "-")
	item := TOCItem{
		Level: 1,
		Title: template.HTML(name),
		Link:  link,
		File:  "_" + file + ".html",
	}
	h.TOC.Sections = append(h.TOC.Sections, &item)
	h.currentTOCItem = &item

	return nil
}

func (h *HTMLWriter) resourceCategoryHeading(title string) template.HTML {
	rendered, err := renderTemplate("resource-category-heading.html", map[string]any{
		"title":     title,
		"sectionID": strings.ToLower(strings.ReplaceAll(title, " ", "-")),
	})
	if err != nil {
		panic(err)
	}

	return rendered
}

func (h *HTMLWriter) sectionHeading(title string) template.HTML {
	rendered, err := renderTemplate("section-heading.html", title)
	if err != nil {
		panic(err)
	}

	return rendered
}

func (h *HTMLWriter) WriteDefinitionsOverview() error {
	if err := writeStaticFile("_definitions.html", h.sectionHeading("Definitions")); err != nil {
		return err
	}

	item := TOCItem{
		Level: 1,
		Title: "DEFINITIONS",
		Link:  "definitions",
		File:  "_definitions.html",
	}
	h.TOC.Sections = append(h.TOC.Sections, &item)
	h.currentTOCItem = &item

	return nil
}

func (h *HTMLWriter) WriteOrphanedOperationsOverview() error {
	if err := writeStaticFile("_operations.html", h.sectionHeading("Operations")); err != nil {
		return err
	}

	item := TOCItem{
		Level: 1,
		Title: "OPERATIONS",
		Link:  "operations",
		File:  "_operations.html",
	}
	h.TOC.Sections = append(h.TOC.Sections, &item)
	h.currentTOCItem = &item

	return nil
}

func (h *HTMLWriter) WriteDefinition(d *api.Definition) error {
	fn := definitionFileName(d)
	path := filepath.Join(api.IncludesDir, fn)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	nvg := fmt.Sprintf("%s %s %s", d.Name, d.Version, d.GroupDisplayName())
	linkID := getLink(nvg)
	title := h.gvkMarkup(d.GroupDisplayName(), d.Version, d.Name)

	// Definitions are added to the TOC to enable the generator to later collect
	// all the individual definition files, but definitions will not show up
	// in the nav treet because it would take up too much screen estate.
	item := TOCItem{
		Level: 2,
		Title: title,
		Link:  linkID,
		File:  fn,
	}
	h.currentTOCItem.SubSections = append(h.currentTOCItem.SubSections, &item)

	return renderTemplateTo(f, "definition.html", map[string]any{
		"nvg":        title,
		"linkID":     linkID,
		"definition": d,
	})
}

func (h *HTMLWriter) WriteOperation(o *api.Operation) error {
	fn := operationFileName(o)
	path := filepath.Join(api.IncludesDir, fn)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	nvg := fmt.Sprintf("%s", o.ID)
	linkID := getLink(nvg)

	oGroup, oVersion, oKind, _ := o.GetGroupVersionKindSub()
	oApiVersion := api.ApiVersion(oVersion)

	title := template.HTML(nvg)
	if len(oGroup) > 0 {
		title = h.gvkMarkup(oGroup, oApiVersion, oKind)
	}

	sort.Slice(o.HttpResponses, func(i, j int) bool {
		return strings.Compare(o.HttpResponses[i].Name, o.HttpResponses[j].Name) < 0
	})

	if err := renderTemplateTo(f, "operation.html", map[string]any{
		"linkID":    linkID,
		"nvg":       nvg,
		"operation": o,
	}); err != nil {
		panic(err)
	}

	item := TOCItem{
		Level: 2,
		Title: title,
		Link:  linkID,
		File:  fn,
	}
	h.currentTOCItem.SubSections = append(h.currentTOCItem.SubSections, &item)

	return nil
}

func (h *HTMLWriter) WriteResource(r *api.Resource) error {
	filename := conceptFileName(r.Definition)
	path := filepath.Join(api.IncludesDir, filename)

	w, err := os.Create(path)
	if err != nil {
		return err
	}
	defer w.Close()

	dvg := fmt.Sprintf("%s %s %s", r.Name, r.Definition.Version, r.Definition.GroupDisplayName())
	linkID := getLink(dvg)

	resourceItem := TOCItem{
		Level: 2,
		Title: h.gvkMarkup(r.Definition.GroupDisplayName(), r.Definition.Version, r.Name),
		Link:  linkID,
		File:  filename,
	}
	h.currentTOCItem.SubSections = append(h.currentTOCItem.SubSections, &resourceItem)

	for _, oc := range r.Definition.OperationCategories {
		if len(oc.Operations) == 0 {
			continue
		}

		ocItem := TOCItem{
			Level: 3,
			Title: template.HTML(oc.Name),
			Link:  oc.TocID(r.Definition),
		}
		resourceItem.SubSections = append(resourceItem.SubSections, &ocItem)

		for _, o := range oc.Operations {
			ocItem.SubSections = append(ocItem.SubSections, &TOCItem{
				Level: 4,
				Title: template.HTML(o.Type.Name),
				Link:  o.TocID(r.Definition),
			})
		}
	}

	if err := renderTemplateTo(w, "resource.html", map[string]any{
		"resource": r,
		"dvg":      resourceItem.Title,
		"linkID":   linkID,
	}); err != nil {
		return err
	}

	return nil
}

func (h *HTMLWriter) WriteOldVersionsOverview() error {
	if err := writeStaticFile("_oldversions.html", h.sectionHeading("Old API Versions")); err != nil {
		return err
	}

	item := TOCItem{
		Level: 1,
		Title: "OLD API VERSIONS",
		Link:  "old-api-versions",
		File:  "_oldversions.html",
	}
	h.TOC.Sections = append(h.TOC.Sections, &item)
	h.currentTOCItem = &item

	return nil
}

func (h *HTMLWriter) Finalize() error {
	if err := os.MkdirAll(api.BuildDir, os.ModePerm); err != nil {
		return err
	}

	if err := h.generateIndex(); err != nil {
		return err
	}

	return nil
}

func (h *HTMLWriter) generateIndex() error {
	html, err := os.Create(filepath.Join(api.BuildDir, "index.html"))
	if err != nil {
		return err
	}
	defer html.Close()

	// collect content from all the individual files we just created
	var content strings.Builder

	collect := func(filename string) {
		fileContent, err := os.ReadFile(filepath.Join(api.IncludesDir, filename))
		if err == nil {
			content.Write(fileContent)
			log.Printf("Collecting %s… \033[32mOK\033[0m", filename)
		} else {
			log.Printf("Collecting %s… \033[31mNot found\033[0m", filename)
		}
	}

	// TODO: Make this a recursive function.
	for _, sec := range h.TOC.Sections {
		collect(sec.File)

		for _, sub := range sec.SubSections {
			if len(sub.File) > 0 {
				collect(sub.File)
			}

			for _, subsub := range sub.SubSections {
				if len(subsub.File) > 0 {
					collect(subsub.File)
				}
			}
		}
	}

	pos := strings.LastIndex(h.Config.SpecVersion, ".")
	release := fmt.Sprintf("release-%s", h.Config.SpecVersion[1:pos])
	specLink := "https://github.com/kubernetes/kubernetes/blob/" + release + "/api/openapi-spec/swagger.json"

	return renderTemplateTo(html, "index.html", map[string]any{
		"toc":      h.TOC,
		"config":   h.Config,
		"specLink": specLink,
		"content":  template.HTML(content.String()),
	})
}

func (h *HTMLWriter) gvkMarkup(group string, version api.ApiVersion, kind string) template.HTML {
	return template.HTML(fmt.Sprintf(`<span class="gvk"><span class="k">%s</span> <span class="v">%s</span> <span class="g">%s</span></span>`, kind, version, group))
}
