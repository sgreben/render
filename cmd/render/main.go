package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"text/template"

	"github.com/sgreben/render/pkg/render"
	"github.com/sirupsen/logrus"
)

var config render.Config
var printVersionFlag bool
var printConfigFlag bool
var printFuncsFlag bool
var version string
var logger *logrus.Entry

func init() {
	logger = logrus.NewEntry(logrus.StandardLogger())

	varsSourcesParameter := varsSourcesParameter{&config.VarsSources}
	varsSourcesFile := varsSourcesFile{&config.VarsSources}
	varsSourcesFileSlurp := varsSourcesFileSlurp{&config.VarsSources}
	varsSourcesEnvPrefix := varsSourcesEnvPrefix{&config.VarsSources}

	templateSourcesParameter := templateSourcesParameter{&config.TemplateSources}
	templateSourcesFile := templateSourcesFile{&config.TemplateSources}
	templateSourcesFileGlob := templateSourcesFileGlob{&config.TemplateSources}

	configPath := configPathParameter{&config}

	flag.Var(&configPath, "config", "path to a config file")

	flag.Var(&varsSourcesParameter, "var", "a single variable definition (<variable>=<value>)")
	flag.Var(&varsSourcesFileSlurp, "var-file-slurp", "set a single variable to a file's contents (or stdin, if - is given) (<variable>=<path>)")
	flag.Var(&varsSourcesFile, "var-file", "load variable values from a file (or stdin, if - is given) ([<key>=]<path>)")
	flag.Var(&varsSourcesEnvPrefix, "var-env", "load variables matching the given prefix from the environment ([<key>=]<prefix>)")

	flag.Var(&templateSourcesParameter, "template", "load a template passed as a parameter ([<template-name>=]<template>)")
	flag.Var(&templateSourcesParameter, "t", "(short for -template)")
	flag.Var(&templateSourcesFile, "template-file", "load a template from a file (or stdin, if - is given) ([<template-name>=]<path>)")
	flag.Var(&templateSourcesFile, "f", "(short for -template-file)")
	flag.Var(&templateSourcesFileGlob, "template-files", "load templates from a set of files matching the given pattern (<glob>)")

	flag.StringVar(&config.ConfigOutPath, "set-config-output-file", "", "path to write the configuration to")
	flag.StringVar(&config.VarsOutPath, "set-vars-output-file", "", "path to write variable values to")
	flag.StringVar(&config.TemplateOutExclude, "set-template-excludes", "", "exclude templates matching the given glob pattern from being output")
	flag.StringVar(&config.TemplateOutPath, "set-output-dir", "", "path to write rendered templates to")
	flag.StringVar(&config.TemplateOutPath, "o", "", "(short for -set-output-dir)")
	flag.StringVar(&config.TemplateLeftDelim, "set-left-delim", "{{", "left template delimiter")
	flag.StringVar(&config.TemplateRightDelim, "set-right-delim", "}}", "right template delimiter")
	flag.StringVar(&config.TemplateOutPrintSeparator, "set-separator", "", "separator template to print between templates when printing templates to stdout")

	flag.BoolVar(&printConfigFlag, "print-config", false, "print config to stdout and exit")
	flag.BoolVar(&config.VarsOutPrint, "print-vars", false, "print variables to stdout and exit")
	flag.BoolVar(&printFuncsFlag, "print-funcs", false, "print available functions and their types to stdout and exit")
	flag.BoolVar(&config.TemplateOutPrint, "print-templates", false, "print rendered templates to stdout")

	flag.BoolVar(&printVersionFlag, "version", false, "print version and exit")
}

func printFuncs(funcs template.FuncMap) {
	maxNameLength := 0
	names := make([]string, len(funcs))
	i := 0
	for name := range funcs {
		if len(name) > maxNameLength {
			maxNameLength = len(name)
		}
		names[i] = name
		i++
	}
	sort.Strings(names)
	format := fmt.Sprintf("%%-%ds %%T\n", maxNameLength)
	for _, name := range names {
		fmt.Printf(format, name, funcs[name])
	}
}

func printVars(vars render.Vars) {
	err := vars.Save(os.Stdout)
	if err != nil {
		logger.WithError(err).Fatal()
	}
}

func printTemplates(templates render.Templates) {
	err := templates.Render(config.TemplateOutExclude, config.TemplateOutPrintSeparator, os.Stdout)
	if err != nil {
		logger.WithError(err).Fatal()
	}
}

func writeVars(vars render.Vars) {
	f, err := os.OpenFile(config.VarsOutPath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		logger.WithError(err).Fatal()
	}
	defer f.Close()
	vars.Save(f)
}

func writeConfig() {
	f, err := os.OpenFile(config.ConfigOutPath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		logger.WithError(err).Fatal()
	}
	defer f.Close()
	err = config.Save(f)
	if err != nil {
		logger.WithError(err).Fatal()
	}
}

func printConfig() {
	err := config.Save(os.Stdout)
	if err != nil {
		logger.WithError(err).Fatal()
	}
}

func main() {
	args := os.Args
	flag.Parse()

	if printVersionFlag {
		fmt.Println(version)
		return
	}

	if config.ConfigOutPath != "" {
		writeConfig()
	}

	if printConfigFlag {
		printConfig()
		return
	}

	funcs := render.Funcs()
	funcs["__RENDER_ARGS"] = func() []string { return args }
	funcs["__RENDER_CONFIG"] = func() render.Config { return config }

	if printFuncsFlag {
		printFuncs(funcs)
		return
	}

	vars := render.Vars{}
	err := vars.FromConfig(&config)
	if err != nil {
		logger.WithError(err).Fatal()
	}

	if config.VarsOutPath != "" {
		writeVars(vars)
	}

	if config.VarsOutPrint {
		printVars(vars)
		return
	}

	templates := render.Templates{Vars: vars}

	// Default behavior: no templates specified -> use stdin
	if len(config.TemplateSources) == 0 {
		config.TemplateSources = []*render.TemplateSource{
			{
				Name:      "stdin",
				FromStdin: &render.TemplateSourceStdin{},
			},
		}
	}

	err = templates.FromConfig(funcs, &config)
	if err != nil {
		logger.WithError(err).Fatal()
	}

	// Default behavior: interpret "-o -" as -print-templates
	if config.TemplateOutPath == "-" {
		config.TemplateOutPath = ""
		config.TemplateOutPrint = true
	}

	if config.TemplateOutPath != "" {
		err = templates.RenderToDir(config.TemplateOutExclude, config.TemplateOutPath)
		if err != nil {
			logger.WithError(err).Fatal()
		}
	} else {
		// Default behavior: no output dir specified -> use stdout
		config.TemplateOutPrint = true
	}

	if config.TemplateOutPrint {
		printTemplates(templates)
	}

}
