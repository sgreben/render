package render

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
)

type VarsSource struct {
	Key            string                `json:",omitempty"`
	FromEnv        *VarsSourceEnv        `json:",omitempty"`
	FromFile       *VarsSourceFile       `json:",omitempty"`
	FromFileSlurp  *VarsSourceFileSlurp  `json:",omitempty"`
	FromFilesSlurp *VarsSourceFilesSlurp `json:",omitempty"`
	FromParameter  *VarsSourceParameter  `json:",omitempty"`
	FromStdin      *VarsSourceStdin      `json:",omitempty"`
}

func (v *VarsSource) Load(vars Vars) error {
	if v.Key != "" {
		if destination, ok := vars[v.Key].(map[string]interface{}); ok {
			vars = Vars(destination)
		} else {
			destination := map[string]interface{}{}
			vars[v.Key] = destination
			vars = Vars(destination)
		}
	}
	if v.FromEnv != nil {
		v.FromEnv.Load(vars)
		return nil
	}
	if v.FromFile != nil {
		return v.FromFile.Load(vars)
	}
	if v.FromFileSlurp != nil {
		return v.FromFileSlurp.Load(vars)
	}
	if v.FromFilesSlurp != nil {
		return v.FromFilesSlurp.Load(vars)
	}
	if v.FromParameter != nil {
		v.FromParameter.Load(vars)
		return nil
	}
	if v.FromStdin != nil {
		return v.FromStdin.Load(vars)
	}
	return nil
}

type VarsSourceParameter struct {
	Key   string
	Value string
}

func (v *VarsSourceParameter) Load(vars Vars) {
	vars[v.Key] = v.Value
}

type VarsSourceFilesSlurp struct {
	Glob string
}

func (v VarsSourceFilesSlurp) Load(vars Vars) error {
	paths, err := filepath.Glob(v.Glob)
	if err != nil {
		return err
	}
	for _, path := range paths {
		err := VarsSourceFileSlurp{Name: path, Path: path}.Load(vars)
		if err != nil {
			return err
		}
	}
	return nil
}

type VarsSourceFileSlurp struct {
	Name string
	Path string
}

func (v VarsSourceFileSlurp) Load(vars Vars) error {
	var bytes []byte
	var err error
	if v.Path == "-" {
		bytes, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
	} else {
		bytes, err = ioutil.ReadFile(v.Path)
		if err != nil {
			return err
		}
	}
	vars[v.Name] = string(bytes)
	return nil
}

type VarsSourceStdin struct{}

func (v VarsSourceStdin) Load(vars Vars) error {
	bs, err := ioutil.ReadAll(os.Stdin)
	buf := bytes.NewBuffer(bs)
	err = vars.fromJSON(buf)
	if err != nil {
		buf = bytes.NewBuffer(bs)
		err = vars.fromYAML(buf)
	}
	if err != nil {
		buf = bytes.NewBuffer(bs)
		err = vars.fromTOML(buf)
	}
	if err != nil {
		return err
	}
	return nil
}

type VarsSourceFile struct {
	Path string
}

func (v VarsSourceFile) Load(vars Vars) error {
	f, err := os.Open(v.Path)
	if err != nil {
		return err
	}
	defer f.Close()
	err = vars.fromJSON(f)
	if err != nil {
		f.Seek(0, os.SEEK_SET)
		err = vars.fromYAML(f)
	}
	if err != nil {
		f.Seek(0, os.SEEK_SET)
		err = vars.fromTOML(f)
	}
	if err != nil {
		return err
	}
	return nil
}

type VarsSourceEnv struct {
	Prefix string `json:",omitempty"`
}

func (v VarsSourceEnv) Load(vars Vars) {
	vars.fromEnv(v.Prefix)
}
