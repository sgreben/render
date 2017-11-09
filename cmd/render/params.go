package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/sgreben/render/pkg/render"
)

type varsSourcesParameter struct {
	store *[]*render.VarsSource
}
type varsSourcesFile struct {
	store *[]*render.VarsSource
}
type varsSourcesFileSlurp struct {
	store *[]*render.VarsSource
}
type varsSourcesEnvPrefix struct {
	store *[]*render.VarsSource
}

func (v *varsSourcesParameter) String() string { return "" }
func (v *varsSourcesFile) String() string      { return "" }
func (v *varsSourcesFileSlurp) String() string { return "" }
func (v *varsSourcesEnvPrefix) String() string { return "" }

func (v *varsSourcesParameter) Set(value string) error {
	i := strings.IndexByte(value, byte('='))
	if i <= 0 {
		return errors.New("syntax: key=value")
	}
	varsSource := &render.VarsSource{
		FromParameter: &render.VarsSourceParameter{
			Key:   value[:i],
			Value: value[i+1:],
		},
	}
	*v.store = append(*v.store, varsSource)
	return nil
}

func (v *varsSourcesFile) Set(value string) error {
	var varsSource *render.VarsSource
	i := strings.IndexByte(value, byte('='))
	key := ""
	if i > 0 {
		key = value[:i]
		value = value[i+1:]
	}
	if value == "-" {
		varsSource = &render.VarsSource{
			Key:       key,
			FromStdin: &render.VarsSourceStdin{},
		}
	} else {
		varsSource = &render.VarsSource{
			Key: key,
			FromFile: &render.VarsSourceFile{
				Path: value,
			},
		}
	}
	*v.store = append(*v.store, varsSource)
	return nil
}

func (v *varsSourcesFileSlurp) Set(value string) error {
	i := strings.IndexByte(value, byte('='))
	if i <= 0 {
		return errors.New("syntax: name=path")
	}
	varsSource := &render.VarsSource{
		FromFileSlurp: &render.VarsSourceFileSlurp{
			Name: value[:i],
			Path: value[i+1:],
		},
	}
	*v.store = append(*v.store, varsSource)
	return nil
}

func (v *varsSourcesEnvPrefix) Set(value string) error {
	i := strings.IndexByte(value, byte('='))
	key := ""
	if i >= 0 {
		key = value[:i]
		value = value[i+1:]
	}
	varsSource := &render.VarsSource{
		Key: key,
		FromEnv: &render.VarsSourceEnv{
			Prefix: value,
		},
	}
	*v.store = append(*v.store, varsSource)
	return nil
}

type templateSourcesParameter struct {
	store *[]*render.TemplateSource
}
type templateSourcesFile struct {
	store *[]*render.TemplateSource
}
type templateSourcesFileGlob struct {
	store *[]*render.TemplateSource
}

func (v *templateSourcesParameter) String() string { return "" }
func (v *templateSourcesFile) String() string      { return "" }
func (v *templateSourcesFileGlob) String() string  { return "" }

func (v *templateSourcesParameter) Set(value string) error {
	i := strings.IndexByte(value, byte('='))
	if i <= 0 {
		name := fmt.Sprintf("__param_%d", len(*v.store))
		i = len(name)
		value = name + "=" + value
	}
	TemplateSource := &render.TemplateSource{
		Name: value[:i],
		FromParameter: &render.TemplateSourceParameter{
			Value: value[i+1:],
		},
	}
	*v.store = append(*v.store, TemplateSource)
	return nil
}

func (v *templateSourcesFile) Set(value string) error {
	i := strings.IndexByte(value, byte('='))
	if i <= 0 {
		i = len(value)
		value = value + "=" + value
	}
	var TemplateSource *render.TemplateSource
	if value == "-" {
		TemplateSource = &render.TemplateSource{
			Name:      value[:i],
			FromStdin: &render.TemplateSourceStdin{},
		}
	} else {
		TemplateSource = &render.TemplateSource{
			Name: value[:i],
			FromFile: &render.TemplateSourceFile{
				Path: value[i+1:],
			},
		}
	}
	*v.store = append(*v.store, TemplateSource)
	return nil
}

func (v *templateSourcesFileGlob) Set(value string) error {
	TemplateSource := &render.TemplateSource{
		Name: value,
		FromFileGlob: &render.TemplateSourceFileGlob{
			Glob: value,
		},
	}
	*v.store = append(*v.store, TemplateSource)
	return nil
}

type configPathParameter struct {
	store *render.Config
}

func (c *configPathParameter) String() string { return "" }
func (c *configPathParameter) Set(value string) error {
	f, err := os.Open(value)
	if err != nil {
		return err
	}
	defer f.Close()
	return c.store.Load(f)
}
