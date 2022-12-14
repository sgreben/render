package render

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/Masterminds/sprig/v3"
	yaml "gopkg.in/yaml.v2"
)

var builtinFuncs = extractBuiltinFuncs()
var sprigFuncs = sprig.TxtFuncMap()
var renderFuncs = template.FuncMap{
	"toCSV": func(value interface{}) (string, error) {
		buf := bytes.NewBuffer([]byte{})
		csv := csv.NewWriter(buf)
		var err error
		if singleRecord, ok := value.([]string); ok {
			err = csv.Write(singleRecord)
			csv.Flush()
		} else if multipleRecords, ok := value.([][]string); ok {
			err = csv.WriteAll(multipleRecords)
			csv.Flush()
		} else {
			err = errors.New("wrong type: must be either []string or [][]string")
		}
		return buf.String(), err
	},
	"fromCSV": func(value string) ([][]string, error) {
		buf := bytes.NewBufferString(value)
		csv := csv.NewReader(buf)
		return csv.ReadAll()
	},
	"toJSON": func(value interface{}) (string, error) {
		bytes, err := json.Marshal(value)
		return string(bytes), err
	},
	"toYAML": func(value interface{}) (string, error) {
		bytes, err := yaml.Marshal(value)
		return string(bytes), err
	},
	"toTOML": func(value interface{}) (string, error) {
		buf := bytes.NewBuffer([]byte{})
		enc := toml.NewEncoder(buf)
		err := enc.Encode(value)
		return buf.String(), err
	},
	"fromJSON": func(value string) (interface{}, error) {
		var obj interface{}
		err := json.Unmarshal([]byte(value), &obj)
		return obj, err
	},
	"fromYAML": func(value string) (interface{}, error) {
		var obj interface{}
		err := yaml.Unmarshal([]byte(value), &obj)
		return obj, err
	},
	"fromTOML": func(value string) (interface{}, error) {
		var obj interface{}
		err := toml.Unmarshal([]byte(value), &obj)
		return obj, err
	},
	"set": func(dict map[string]interface{}, kvs ...interface{}) map[string]interface{} {
		for i := 0; i < len(kvs); i += 2 {
			dict[kvs[i].(string)] = kvs[i+1]
		}
		return dict
	},
	"unset": func(dict map[string]interface{}, keys ...string) map[string]interface{} {
		for _, key := range keys {
			delete(dict, key)
		}
		return dict
	},
}

func asEmptyInterfaceSlice(obj interface{}) []interface{} {
	if slice, ok := obj.([]interface{}); ok {
		return slice
	}
	value := reflect.ValueOf(obj)
	slice := make([]interface{}, value.Len())
	for i := 0; i < value.Len(); i++ {
		slice[i] = value.Index(i).Interface()
	}
	return slice
}

func reflectMap(f reflect.Value, fArgs []interface{}, slice []interface{}, lastArg bool) ([]interface{}, error) {
	output := make([]interface{}, len(slice))
	argsValues := make([]reflect.Value, 1+len(fArgs))
	if lastArg {
		for i, arg := range fArgs {
			argsValues[i] = reflect.ValueOf(arg)
		}
	} else {
		for i, arg := range fArgs {
			argsValues[1+i] = reflect.ValueOf(arg)
		}
	}
	for i, x := range slice {
		if lastArg {
			argsValues[len(argsValues)-1] = reflect.ValueOf(x)
		} else {
			argsValues[0] = reflect.ValueOf(x)
		}
		returnValues := f.Call(argsValues)
		if len(returnValues) == 2 {
			if err, ok := returnValues[1].Interface().(error); ok && err != nil {
				return nil, err
			}
		}
		output[i] = returnValues[0].Interface()
	}
	return output, nil
}

func reflectFilter(f reflect.Value, fArgs []interface{}, slice []interface{}, lastArg bool) ([]interface{}, error) {
	output := make([]interface{}, 0)
	argsValues := make([]reflect.Value, 1+len(fArgs))
	if lastArg {
		for i, arg := range fArgs {
			argsValues[i] = reflect.ValueOf(arg)
		}
	} else {
		for i, arg := range fArgs {
			argsValues[1+i] = reflect.ValueOf(arg)
		}
	}
	for _, x := range slice {
		if lastArg {
			argsValues[len(argsValues)-1] = reflect.ValueOf(x)
		} else {
			argsValues[0] = reflect.ValueOf(x)
		}
		returnValues := f.Call(argsValues)
		if len(returnValues) == 2 {
			if err, ok := returnValues[1].Interface().(error); ok && err != nil {
				return nil, err
			}
		}
		keep := returnValues[0].Interface()
		if keep, ok := keep.(bool); ok && keep {
			output = append(output, x)
		}
	}
	return output, nil
}

func addHigherOrderFuncs(funcs template.FuncMap, builtinFuncs template.FuncMap) {
	findFunc := func(funcName string) (reflect.Value, error) {
		f, ok := funcs[funcName]
		if !ok {
			f, ok = builtinFuncs[funcName]
			if !ok {
				return reflect.Value{}, fmt.Errorf("no such function: '%s'", funcName)
			}
		}
		return reflect.ValueOf(f), nil
	}

	funcs["map"] = func(funcName string, mapArgs ...interface{}) ([]interface{}, error) {
		f, err := findFunc(funcName)
		if err != nil {
			return nil, err
		}
		n := len(mapArgs)
		return reflectMap(f, mapArgs[:n-1], asEmptyInterfaceSlice(mapArgs[n-1]), false)
	}

	funcs["filter"] = func(funcName string, filterArgs ...interface{}) ([]interface{}, error) {
		f, err := findFunc(funcName)
		if err != nil {
			return nil, err
		}
		n := len(filterArgs)
		return reflectFilter(f, filterArgs[:n-1], asEmptyInterfaceSlice(filterArgs[n-1]), false)
	}

	funcs["mapFlip"] = func(funcName string, mapArgs ...interface{}) ([]interface{}, error) {
		f, err := findFunc(funcName)
		if err != nil {
			return nil, err
		}
		n := len(mapArgs)
		return reflectMap(f, mapArgs[:n-1], asEmptyInterfaceSlice(mapArgs[n-1]), true)
	}

	funcs["filterFlip"] = func(funcName string, filterArgs ...interface{}) ([]interface{}, error) {
		f, err := findFunc(funcName)
		if err != nil {
			return nil, err
		}
		n := len(filterArgs)
		return reflectFilter(f, filterArgs[:n-1], asEmptyInterfaceSlice(filterArgs[n-1]), true)
	}
}

func extractBuiltinFuncs() template.FuncMap {
	funcs := map[string]interface{}{}
	type callData struct {
		Args        []interface{}
		ReturnValue interface{}
	}
	returnFuncs := func(callData *callData) template.FuncMap {
		return template.FuncMap{
			"return": func(obj interface{}) interface{} {
				callData.ReturnValue = obj
				return obj
			},
		}
	}
	callBuiltinFunc := func(name string) interface{} {
		return func(args ...interface{}) (interface{}, error) {
			templateString := fmt.Sprintf(`{{return (%s`, name)
			for i := 0; i < len(args); i++ {
				templateString += fmt.Sprintf(" (index .Args %d)", i)
			}
			templateString += ")}}"
			callData := callData{Args: args}
			err := template.Must(template.New("").Funcs(returnFuncs(&callData)).Parse(templateString)).Execute(ioutil.Discard, callData)
			return callData.ReturnValue, err
		}
	}
	names := []string{
		"and",
		"call",
		"html",
		"index",
		"js",
		"len",
		"not",
		"or",
		"print",
		"printf",
		"println",
		"urlquery",
		"eq",
		"ge",
		"gt",
		"le",
		"lt",
		"ne",
	}
	for _, name := range names {
		funcs[name] = callBuiltinFunc(name)
	}
	return funcs
}

func Funcs() template.FuncMap {
	funcs := sprigFuncs
	delete(funcs, "hello")
	delete(funcs, "toJson")
	delete(funcs, "toPrettyJson")
	delete(funcs, "env")
	delete(funcs, "expandenv")
	for key, value := range renderFuncs {
		funcs[key] = value
	}
	addHigherOrderFuncs(funcs, builtinFuncs)
	return funcs
}
