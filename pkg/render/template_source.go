package render

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

type TemplateSource struct {
	Name          string
	FromEnv       *TemplateSourceEnv       `json:",omitempty"`
	FromFile      *TemplateSourceFile      `json:",omitempty"`
	FromFileGlob  *TemplateSourceFileGlob  `json:",omitempty"`
	FromParameter *TemplateSourceParameter `json:",omitempty"`
	FromStdin     *TemplateSourceStdin     `json:",omitempty"`
}

func (ts *TemplateSource) Load(funcs template.FuncMap, t *template.Template) ([]string, error) {
	if ts.FromEnv != nil {
		return ts.FromEnv.Load(funcs, ts.Name, t)
	}
	if ts.FromFile != nil {
		return ts.FromFile.Load(funcs, ts.Name, t)
	}
	if ts.FromFileGlob != nil {
		return ts.FromFileGlob.Load(funcs, ts.Name, t)
	}
	if ts.FromParameter != nil {
		return ts.FromParameter.Load(funcs, ts.Name, t)
	}
	if ts.FromStdin != nil {
		return ts.FromStdin.Load(funcs, ts.Name, t)
	}
	return nil, nil
}

type TemplateSourceParameter struct {
	Value string
}

func (ts *TemplateSourceParameter) Load(funcs template.FuncMap, name string, t *template.Template) ([]string, error) {
	t = t.New(name).Funcs(funcs)
	setupTemplate(t)
	_, err := t.Parse(ts.Value)
	return []string{name}, err
}

type TemplateSourceStdin struct{}

func (ts *TemplateSourceStdin) Load(funcs template.FuncMap, name string, t *template.Template) ([]string, error) {
	bytes, err := ioutil.ReadAll(os.Stdin)
	t = t.New(name).Funcs(funcs)
	setupTemplate(t)
	_, err = t.Parse(string(bytes))
	return []string{name}, err
}

type TemplateSourceFileGlob struct {
	Glob string
}

func (ts *TemplateSourceFileGlob) Load(funcs template.FuncMap, name string, t *template.Template) ([]string, error) {
	paths, err := filepath.Glob(ts.Glob)
	if err != nil {
		return nil, err
	}
	for _, path := range paths {
		tsf := &TemplateSourceFile{Path: path}
		_, err := tsf.Load(funcs, path, t)
		if err != nil {
			return nil, err
		}
	}
	return paths, nil
}

type TemplateSourceFile struct {
	Path string
}

func (ts *TemplateSourceFile) Load(funcs template.FuncMap, name string, t *template.Template) ([]string, error) {
	f, err := os.Open(ts.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	bytes, err := ioutil.ReadFile(ts.Path)
	if err != nil {
		return nil, err
	}

	t = t.New(name).Funcs(funcs)
	setupTemplate(t)
	_, err = t.Parse(string(bytes))
	return []string{name}, err
}

type TemplateSourceEnv struct {
	Key string
}

func (ts *TemplateSourceEnv) Load(funcs template.FuncMap, name string, t *template.Template) ([]string, error) {
	t = t.New(name).Funcs(funcs)
	setupTemplate(t)
	_, err := t.Parse(os.Getenv(ts.Key))
	return []string{name}, err
}
