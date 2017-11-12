package render

import (
	"io"
	"os"
	"path"
	"text/template"

	"github.com/gobwas/glob"
)

type Templates struct {
	Root    *template.Template
	Funcs   template.FuncMap
	Vars    map[string]interface{}
	Names   []string
	Exclude glob.Glob
}

func setupTemplate(t *template.Template) {
	t.Option("missingkey=zero")
}

func (t *Templates) Render(excludes string, separator string, w io.Writer) error {
	separatorTemplate := template.New("separator")
	separatorTemplate.Funcs(t.Funcs)
	setupTemplate(separatorTemplate)
	_, err := separatorTemplate.Parse(separator)
	if err != nil {
		return err
	}

	n := len(t.Names)
	for i, templateName := range t.Names {
		template := t.Root.Lookup(templateName)
		if t.Exclude != nil && t.Exclude.Match(template.Name()) {
			continue
		}
		err = template.Execute(w, t.Vars)
		if err != nil {
			return err
		}
		if i < n-1 {
			err = separatorTemplate.Execute(w, t.Vars)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *Templates) RenderToDir(excludes string, dir string) error {
	for _, templateName := range t.Names {
		template := t.Root.Lookup(templateName)
		if t.Exclude != nil && t.Exclude.Match(template.Name()) {
			continue
		}
		templatePath := path.Join(dir, template.Name())
		templatePathDir := path.Dir(templatePath)
		err := os.MkdirAll(templatePathDir, 0777|os.ModeDir)
		if err != nil {
			return err
		}
		os.Remove(templatePath)
		f, err := os.OpenFile(templatePath, os.O_CREATE|os.O_WRONLY, 0777)
		if err != nil {
			return err
		}
		func() {
			defer f.Close()
			err = template.Execute(f, t.Vars)
		}()
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Templates) FromConfig(funcs template.FuncMap, config *Config) error {
	exclude, err := glob.Compile(config.TemplateOutExclude)
	if err != nil {
		return err
	}
	t.Exclude = exclude
	t.Root = template.New("root")
	t.Root.Delims(config.TemplateLeftDelim, config.TemplateRightDelim)
	t.Funcs = funcs
	t.Names = []string{}
	for _, templateSource := range config.TemplateSources {
		names, err := templateSource.Load(funcs, t.Root)
		if err != nil {
			return err
		}
		t.Names = append(t.Names, names...)
	}
	return nil
}
