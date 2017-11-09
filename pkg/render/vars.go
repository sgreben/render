package render

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	toml "github.com/BurntSushi/toml"
	yaml "gopkg.in/yaml.v2"
)

// Vars is the context in which templates are executed
type Vars map[string]interface{}

func (v Vars) Save(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	err := enc.Encode(v)
	if err != nil {
		bytes, err := yaml.Marshal(v)
		if err != nil {
			return err
		}
		_, err = w.Write(bytes)
		return err
	}
	return err
}

func (v Vars) overwriteWith(other Vars) {
	for key, value := range other {
		v[key] = value
	}
}

// The YAML decoder generates map[interface{}]interface{} even
// when the key type is string. This function scans for such
// overly-generic map types and narrows them.
func flatten(value interface{}) interface{} {

	if m, ok := value.(Vars); ok {
		for k, v := range m {
			m[k] = flatten(v)
		}
		return map[string]interface{}(m)
	}

	if m, ok := value.(map[string]interface{}); ok {
		for k, v := range m {
			m[k] = flatten(v)
		}
		return m
	}

	if m, ok := value.(map[interface{}]interface{}); ok {
		allKeysAreStringsOrInts := true
		for k := range m {
			if _, ok := k.(string); !ok {
				if _, ok := k.(int); !ok {
					allKeysAreStringsOrInts = false
				}
			}
		}
		if allKeysAreStringsOrInts {
			mVars := map[string]interface{}{}
			for k, v := range m {
				if k, ok := k.(string); ok {
					mVars[k] = flatten(v)
				}
				if k, ok := k.(int); ok {
					mVars[strconv.Itoa(k)] = flatten(v)
				}
			}
			return mVars
		}
		return m
	}

	if s, ok := value.([]interface{}); ok {
		for i, v := range s {
			s[i] = flatten(v)
		}
	}
	return value
}

func (v Vars) fromYAML(r io.Reader) error {
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	raw := Vars{}
	err = yaml.Unmarshal(bytes, &raw)
	if err != nil {
		return err
	}
	v.overwriteWith(flatten(raw).(map[string]interface{}))
	return nil
}

func (v Vars) fromTOML(r io.Reader) error {
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	raw := Vars{}
	err = toml.Unmarshal(bytes, &raw)
	if err != nil {
		return err
	}
	v.overwriteWith(flatten(raw).(map[string]interface{}))
	return nil
}

func (v Vars) fromJSON(r io.Reader) error {
	vars := Vars{}
	dec := json.NewDecoder(r)
	err := dec.Decode(&vars)
	if err != nil {
		return err
	}
	v.overwriteWith(vars)
	return nil
}

func (v Vars) fromEnvSingle(key string) {
	v[key] = os.Getenv(key)
}

func (v Vars) fromEnv(prefix string) {
	for _, entry := range os.Environ() {
		i := strings.IndexByte(entry, byte('='))
		key := entry[:i]
		if strings.HasPrefix(key, prefix) {
			value := entry[i+1:]
			v[key] = value
		}
	}
}

func (v Vars) FromConfig(config *Config) error {
	for _, varsSource := range config.VarsSources {
		err := varsSource.Load(v)
		if err != nil {
			return err
		}
	}
	return nil
}
