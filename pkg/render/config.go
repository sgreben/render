package render

import (
	"encoding/json"
	"io"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// Config is the run-time configuration of the app
type Config struct {
	ConfigOutPath             string            `json:",omitempty"`
	TemplateOutExclude        string            `json:",omitempty"`
	TemplateOutPrintSeparator string            `json:",omitempty"`
	TemplateOutPrint          bool              `json:",omitempty"`
	TemplateOutPath           string            `json:",omitempty"`
	TemplateLeftDelim         string            `json:",omitempty"`
	TemplateRightDelim        string            `json:",omitempty"`
	TemplateSources           []*TemplateSource `json:",omitempty"`
	VarsOutPrint              bool              `json:",omitempty"`
	VarsOutPath               string            `json:",omitempty"`
	VarsSources               []*VarsSource     `json:",omitempty"`
}

func (c *Config) Save(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	err := enc.Encode(c)
	if err != nil {
		bytes, err := yaml.Marshal(c)
		if err != nil {
			return err
		}
		_, err = w.Write(bytes)
		return err
	}
	return err
}

func (c *Config) Load(r io.Reader) error {
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, &c)
	return err
	if err == nil {
		return nil
	}
	return yaml.Unmarshal(bytes, &c)
}
