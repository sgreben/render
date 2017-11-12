# render

[![CircleCI](https://circleci.com/gh/sgreben/render.svg?style=svg)](https://circleci.com/gh/sgreben/render) [![Docker Repository on Quay](https://quay.io/repository/sergey_grebenshchikov/render/status "Docker Repository on Quay")](https://quay.io/repository/sergey_grebenshchikov/render)

A flexible go-template renderer.

```bash
docker pull quay.io/sergey_grebenshchikov/render
```

```bash
go get -u github.com/sgreben/render/cmd/render
```

- Template definitions can be given as command-line arguments (`-template`, `-t`), or read from files (`-template-file`, `-f`, `-template-files`).
- Variable definitions can be given as command-line arguments (`-var`), taken from the environment (`-var-env`), or read from JSON / YAML / TOML files (`-var-file`).

The template syntax is described at <https://golang.org/pkg/text/template>

## Examples

```bash
$ render -var foo=bar -t "value of foo: {{ .foo }}"
value of foo: bar
```

```bash
$ echo "{{ .SHELL }} {{ .USER }}" | render -var-env ""
/bin/bash sgreben
```

```bash
$ echo "{{ .env.SHELL }} {{ .env.USER }}" | render -var-env env=
/bin/bash sgreben
```

```bash
$ echo '{ "foo": { "bar": "baz" } }' | render -var-file - -t "value of foo.bar: {{ .foo.bar }}"
value of foo.bar: baz
```

```bash
$ render -t "{{ __RENDER_ARGS | toJSON }}"
["render","-t","{{ __RENDER_ARGS | toJSON }}"]
```

A longer example demonstrating usage of multiple templates, saving CLI flags as a reusable config, and rendering templates to disk:

```bash
$ find . -type f
./components/containers
./components/volumes
./templates/pod.yml
./vars.yml

$ cat templates/pod.yml
apiVersion: v1
kind: Pod
metadata:
  name: {{ .name }}
spec:
  containers: {{ template "components/containers" . }}
  volumes: {{ template "components/volumes" . }

$ cat components/containers
{{- $volumeMounts := .volumes | map "pick" "name" "mountPath" -}}
{{- .containers | map "set" "volumeMounts" $volumeMounts | toJSON -}}

$ cat components/volumes
{{ .volumes | map "omit" "mountPath" | toJSON }}

$ cat vars.yml
name: my-pod
containers:
- name: nginx
  image: nginx:latest
- name: busybox
  image: busybox
volumes:
- name: data
  mountPath: /data
  emptyDir: {}
- name: config
  mountPath: /config
  configMap:
    name: my-configmap

# Generate a config file for render
$ render -var-file vars.yml -template-files 'components/*' -template-files 'templates/*' -set-template-excludes 'components/*' -o rendered -print-config > config.json

$ cat config.json
{
  "TemplateOutExclude": "components/*",
  "TemplateOutPath": "rendered",
  "TemplateLeftDelim": "{{",
  "TemplateRightDelim": "}}",
  "TemplateSources": [
    {
      "Name": "components/*",
      "FromFileGlob": {
        "Glob": "components/*"
      }
    },
    {
      "Name": "templates/*",
      "FromFileGlob": {
        "Glob": "templates/*"
      }
    }
  ],
  "VarsSources": [
    {
      "FromFile": {
        "Path": "vars.yml"
      }
    }
  ]
}

# Render the templates using the config file
$ render -config config.json

$ find . -type f
./components/containers
./components/volumes
./config.json
./rendered/templates/pod.yml # this is new
./templates/pod.yml
./vars.yml

$ cat rendered/templates/pod.yml
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers: [{"image":"nginx:latest","name":"nginx","volumeMounts":[{"mountPath":"/data","name":"data"},{"mountPath":"/config","name":"config"}]},{"image":"busybox","name":"busybox","volumeMounts":[{"mountPath":"/data","name":"data"},{"mountPath":"/config","name":"config"}]}]
  volumes: [{"emptyDir":{},"name":"data"},{"configMap":{"name":"my-configmap"},"name":"config"}]
```

## Tips

- Variable definitions are applied in the order in which they are given on the command line / in the config file. Later definitions of the same variable override its earlier definitions.
- Templates are loaded and rendered in the order in which they are given on the command line / in the config file. If templates with the same name are given, later definitions override earlier definitions.
- If no template input flags are given, `render` defaults to reading a template from `stdin`.
- If no template output flags are given, `render` defaults to rendering templates to `stdout`.

## Template functions

Additionally to the functions listed at <https://golang.org/pkg/text/template/#hdr-Functions>, and the functions provided by the [Sprig library](https://godoc.org/github.com/Masterminds/sprig) (except `env` and `expandenv`), the following functions are defined:

- `toCSV`
- `fromCSV`
- `toJSON`
- `fromJSON`
- `toYAML`
- `fromYAML`
- `toTOML`
- `fromTOML`
- `map`
    ```go-template
    pipeline | map "functionName" arg0 arg1 ...
    ```

    produces the list

    ```go-template
    {
      functionName(pipeline[0], arg0, arg1, ...),
      functionName(pipeline[1], arg0, arg1, ...),
      functionName(pipeline[2], arg0, arg1, ...),
      ...
    }
    ```
- `mapFlip` -- same as `map`, except `pipeline[i]` is provided as the *last* argument to `functionName`, not as the first.
- `filter`
    ```go-template
    pipeline | filter "functionName" arg0 arg1 ...
    ```
    produces the list of all `pipeline[i]` for which `functionName(pipeline[i], arg0, arg1, ...)` returns `true`
- `filterFlip` -- same as `filter`, except `pipeline[i]` is provided as the *last* argument to `functionName`, not as the first.
- `__RENDER_ARGS`
- `__RENDER_CONFIG`

To obtain a list of all defined functions, run `render -print-funcs`.

## Build

- Binary

    ```bash
    dep ensure
    make bin/render
    ```

- Docker image

    ```bash
    make build
    ```

## Usage

```text
Usage of render:
  -config value
    	path to a config file
  -f value
    	(short for -template-file)
  -o string
    	(short for -set-output-dir)
  -print-config
    	print config to stdout and exit
  -print-funcs
    	print available functions and their types to stdout and exit
  -print-templates
    	print rendered templates to stdout
  -print-vars
    	print variables to stdout and exit
  -set-config-output-file string
    	path to write the configuration to
  -set-left-delim string
    	left template delimiter (default "{{")
  -set-output-dir string
    	path to write rendered templates to
  -set-right-delim string
    	right template delimiter (default "}}")
  -set-separator string
    	separator template to print between templates when printing templates to stdout
  -set-template-excludes string
    	exclude templates matching the given glob pattern from being output
  -set-vars-output-file string
    	path to write variable values to
  -t value
    	(short for -template)
  -template value
    	load a template passed as a parameter ([<template-name>=]<template>)
  -template-file value
    	load a template from a file (or stdin, if - is given) ([<template-name>=]<path>)
  -template-files value
    	load templates from a set of files matching the given pattern (<glob>)
  -var value
    	a single variable definition (<variable>=<value>)
  -var-env value
    	load variables matching the given prefix from the environment ([<key>=]<prefix>)
  -var-file value
    	load variable values from a file (or stdin, if - is given) ([<key>=]<path>)
  -var-file-slurp value
    	set a single variable to a file's contents (or stdin, if - is given) (<variable>=<path>)
  -var-files-slurp value
    	load all files matching the given glob pattern as variables ([<key>]=<glob>)
  -version
    	print version and exit
```