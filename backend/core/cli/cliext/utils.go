package cliext

import (
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
	"reflect"
	"strconv"
	"strings"
)

func DecodeCliFlagsTo(ctx *cli.Context, name string, v interface{}) error {
	var ptrRef reflect.Value
	if ref, ok := v.(reflect.Value); ok {
		ptrRef = ref
	} else {
		ptrRef = reflect.ValueOf(v)
	}

	if ptrRef.Kind() == reflect.Ptr {
		//初始化空指针
		if ptrRef.IsNil() && ptrRef.CanSet() {
			ptrRef.Set(reflect.New(ptrRef.Type().Elem()))
		}
		ptrRef = ptrRef.Elem()
	}

	for i := 0; i < ptrRef.NumField(); i++ {
		fv := ptrRef.Field(i)
		ft := ptrRef.Type().Field(i)
		if fv.Kind() == reflect.Ptr {
			fv = ptrRef.Elem()
		}
		yamlTag := ft.Tag.Get("yaml")
		var commandName string
		if len(yamlTag) > 0 {
			commandName = name + "." + yamlTag
		} else {
			commandName = name + "." + ft.Name
		}
		commandName = strings.ToLower(strings.TrimLeft(commandName, "."))

		if typeValue, ok := ft.Tag.Lookup("type"); ok {
			switch typeValue {
			case "enum":
				if !ok {
					return errors.New("Error")
				}
				value := ctx.Generic(commandName)
				enumValue := value.(*EnumValue)
				selected := enumValue.selected
				var index int64 = -1
				for k, v := range enumValue.Enum {
					if v == selected {
						index = int64(k)
						break
					}
				}

				fv.SetInt(index)
				continue
			}
		}

		switch fv.Kind() {
		case reflect.Struct:
			err := DecodeCliFlagsTo(ctx, commandName, fv)
			if err != nil {
				return err
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if !fv.CanSet() {
				continue
			}
			value := ctx.Int64(commandName)
			fv.SetInt(value)
		case reflect.Uint64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
			if !fv.CanSet() {
				continue
			}
			value := ctx.Uint64(commandName)
			fv.SetUint(value)
		case reflect.Bool:
			if !fv.CanSet() {
				continue
			}
			value := ctx.Bool(commandName)
			fv.SetBool(value)
		default:
			if !fv.CanSet() {
				continue
			}
			value := ctx.String(commandName)
			fv.SetString(value)
		}
	}

	return nil
}
func GenerateCliFlags(v interface{}, prefix, name string, flags *[]cli.Flag) (err error) {
	var ptrRef reflect.Value
	if ref, ok := v.(reflect.Value); ok {
		ptrRef = ref
	} else {
		ptrRef = reflect.ValueOf(v)
	}
	if ptrRef.Kind() == reflect.Ptr {
		ptrRef = ptrRef.Elem()
	}
	for i := 0; i < ptrRef.NumField(); i++ {
		fv := ptrRef.Field(i)
		ft := ptrRef.Type().Field(i)

		if fv.Kind() == reflect.Ptr {
			fv = ptrRef.Elem()
		}
		yamlTag := ft.Tag.Get("yaml")
		var commandName string
		if len(yamlTag) > 0 {
			commandName = name + "." + yamlTag
		} else {
			commandName = name + "." + ft.Name
		}
		commandName = strings.ToLower(strings.TrimLeft(commandName, "."))
		envName := strcase.ToScreamingSnake(strings.Replace(prefix+"."+commandName, ".", "_", -1))
		defaultValue := ft.Tag.Get("default")
		required := false
		_, required = ft.Tag.Lookup("required")

		if typeValue, ok := ft.Tag.Lookup("type"); ok {
			switch typeValue {
			case "enum":
				if _, ok = fv.Type().MethodByName("Values"); !ok {
					return errors.New(fmt.Sprintf("Field: %s Type: %s  should be have returns method return all strings returns", ft.Name, ft.Type.Name()))
				}
				method := fv.MethodByName("Values")
				//if !ok{
				//}
				//spew.Dump(runtime.EnvMode(1).Values())
				returns := method.Call(nil)

				first := returns[0]
				if !first.CanInterface(){
					return errors.New(fmt.Sprintf("Field: %s Type: %s  Values method return can not convert Interface", ft.Name, ft.Type.Name()))
				}
				values,ok:=first.Interface().([]string)
				if !ok {
					return errors.New(fmt.Sprintf("Field: %s Type: %s  Values method return can not convert []string", ft.Name, ft.Type.Name()))
				}
				enums := make([]string, 0, len(values))
				for _, v := range values {
					enums = append(enums, strcase.ToSnake(v))
				}

				*flags = append(*flags, &cli.GenericFlag{
					Name:    commandName,
					Aliases: nil,
					Usage:   "",
					EnvVars: []string{
						envName,
					},
					Required:  required,
					Hidden:    false,
					TakesFile: false,
					Value: &EnumValue{
						Enum:    enums,
						Default: defaultValue,
					},
				})

				continue
			}
		}

		switch fv.Kind() {
		case reflect.Struct:
			err := GenerateCliFlags(fv, prefix, commandName, flags)
			if err != nil {
				return err
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
			reflect.Uint8, reflect.Uint16, reflect.Uint32:
			value := 0
			if defaultValue != "" {
				value, err = strconv.Atoi(defaultValue)
				if err != nil {
					return err
				}

			}
			*flags = append(*flags, altsrc.NewIntFlag(&cli.IntFlag{
				Name:  commandName,
				Value: value,
				EnvVars: []string{
					envName,
				},
				Required: required,
			}))
		case reflect.Uint64, reflect.Uint:
			value, err := strconv.ParseUint(defaultValue, 10, 10)
			if err != nil {
				return err
			}
			*flags = append(*flags, altsrc.NewUint64Flag(&cli.Uint64Flag{
				Name:  name,
				Value: value,
				EnvVars: []string{
					envName,
				},
				Required: required,
			}))
		case reflect.Int64:
			value, err := strconv.ParseInt(defaultValue, 10, 10)
			if err != nil {
				return err
			}
			*flags = append(*flags, altsrc.NewInt64Flag(&cli.Int64Flag{
				Name:  commandName,
				Value: value,
				EnvVars: []string{
					envName,
				},
				Required: required,
			}))
		case reflect.Bool:
			var boolValue bool
			defaultValue = strings.ToLower(defaultValue)
			if defaultValue == "true" {
				boolValue = true
			}
			*flags = append(*flags, altsrc.NewBoolFlag(&cli.BoolFlag{
				Name:  commandName,
				Value: boolValue,
				EnvVars: []string{
					envName,
				},
				Required: required,
			}))
		default:

			*flags = append(*flags, altsrc.NewStringFlag(&cli.StringFlag{
				Name:  commandName,
				Value: defaultValue,
				EnvVars: []string{
					envName,
				},
				Required: required,
			}))
		}

	}

	return nil
}
