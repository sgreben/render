package render

import (
	"path/filepath"

	"github.com/gobwas/glob"
)

type Files map[string]interface{}

func (f Files) GetBytes(name string) []byte {
	return []byte(f.Get(name))
}

func (f Files) Get(name string) string {
	return f[name].(string)
}

func (f Files) Glob(pattern string) (Files, error) {
	g, err := glob.Compile(pattern, filepath.Separator)
	if err != nil {
		return nil, err
	}
	result := Files{}
	for name := range f {
		if g.Match(name) {
			result[name] = f[name]
		}
	}
	return result, nil
}
