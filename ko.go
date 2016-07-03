package ko

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/kovetskiy/toml"
)

type (
	// Unmarshaller represents signature of function that will be used for
	// unmarshalling raw file data to structured data.
	// See:
	//   json.Unmarshal
	//   toml.Unmarshal
	Unmarshaller func([]byte, interface{}) error
)

var (
	// DefaultUnmarshaller will be used for unmarshalling if no unmarshaller
	// specified.
	DefaultUnmarshaller Unmarshaller = toml.Unmarshal
)

// Load resource data from specified file. unmarshaller variable can be passed
// if you want to use custom unmarshaller, by default will be used
// DefaultUnmarshaller (toml.Unmarshal)
func Load(
	path string,
	resource interface{},
	unmarshaller ...Unmarshaller,
) error {
	if len(unmarshaller) > 1 {
		panic("passed more then one unmarshaller")
	}

	if len(unmarshaller) == 0 {
		unmarshaller = append(unmarshaller, DefaultUnmarshaller)
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = unmarshaller[0](data, resource)
	if err != nil {
		return err
	}

	err = validate(resource)
	if err != nil {
		return err
	}

	return nil
}

func validate(
	value interface{},
	prefix ...string,
) error {
	resource := reflect.Indirect(reflect.ValueOf(value))
	if resource.Kind() == reflect.Map {
		return nil
	}

	if resource.Kind() != reflect.Struct {
		return fmt.Errorf("resource should be a struct")
	}

	resourceStruct := resource.Type()
	for index := 0; index < resourceStruct.NumField(); index++ {
		var (
			resourceField = resource.Field(index)
			structField   = resourceStruct.Field(index)
		)

		if reflect.DeepEqual(
			resourceField.Interface(),
			reflect.Zero(resourceField.Type()).Interface(),
		) {
			defaultValue := structField.Tag.Get("default")
			if defaultValue != "" {
				err := yaml.Unmarshal(
					[]byte(defaultValue),
					resourceField.Addr().Interface(),
				)
				if err != nil {
					return err
				}
			} else if structField.Tag.Get("required") == "true" {
				return fmt.Errorf(
					"%s is required, but no value specified",
					strings.Join(append(prefix, structField.Name), "."),
				)
			}
		}

		for resourceField.Kind() == reflect.Ptr {
			resourceField = resourceField.Elem()
		}

		if resourceField.Kind() == reflect.Struct {
			err := validate(
				resourceField.Addr().Interface(),
				append(prefix, structField.Name)...,
			)
			if err != nil {
				return err
			}

			continue
		}

		if resourceField.Kind() == reflect.Slice {
			for i := 0; i < resourceField.Len(); i++ {
				field := reflect.Indirect(resourceField.Index(i))
				if field.Kind() == reflect.Struct {
					err := validate(
						field.Addr().Interface(),
						append(
							prefix, fmt.Sprintf("%s[%d]", structField.Name, i),
						)...,
					)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}
