package ko

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/iancoleman/strcase"
	"github.com/reconquest/karma-go"
	"gopkg.in/yaml.v3"
)

type (
	// Unmarshaller represents signature of function that will be used for
	// unmarshalling raw file data to structured data.
	// See:
	//   json.Unmarshal
	//   toml.Unmarshal
	Unmarshaller func([]byte, interface{}) error
)

// DefaultUnmarshaller will be used for unmarshalling if no unmarshaller
// specified.
var DefaultUnmarshaller Unmarshaller = toml.Unmarshal

type (
	// RequireFile is an option for Load method which can be used to skip
	// non-existing file and load all default values for the config fields.
	RequireFile bool
)

// Load resource data from specified file. unmarshaller variable can be passed
// if you want to use custom unmarshaller, by default will be used
// DefaultUnmarshaller (toml.Unmarshal)
func Load(
	path string,
	resource interface{},
	opts ...interface{},
) error {
	var unmarshaller Unmarshaller
	var requireFile bool = true

	for _, opt := range opts {
		switch opt := opt.(type) {
		case func([]byte, interface{}) error:
			unmarshaller = opt
		case Unmarshaller:
			unmarshaller = opt
		case RequireFile:
			requireFile = bool(opt)
		}
	}

	if unmarshaller == nil {
		unmarshaller = DefaultUnmarshaller
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) && requireFile {
			return err
		}
	}

	err = unmarshaller(data, resource)
	if err != nil {
		return err
	}

	err = validate(resource, true)
	if err != nil {
		return err
	}

	return nil
}

func validate(
	value interface{},
	parentRequired bool,
	prefix ...string,
) error {
	resource := reflect.Indirect(reflect.ValueOf(value))
	if resource.Kind() == reflect.Map {
		return nil
	}

	if resource.Kind() == reflect.Slice {
		for i := 0; i < resource.Len(); i++ {
			err := validate(resource.Index(i), parentRequired, prefix...)
			if err != nil {
				return karma.Format(
					err,
					"%d item is invalid",
					i,
				)
			}
		}

		return nil
	}

	if resource.Kind() != reflect.Struct {
		return fmt.Errorf("resource should be a struct")
	}

	resourceStruct := resource.Type()
	for index := 0; index < resourceStruct.NumField(); index++ {
		var (
			resourceField       = resource.Field(index)
			structField         = resourceStruct.Field(index)
			fieldName           = string(structField.Name)
			structFieldRequired = structField.Tag.Get("required") == "true"
		)

		if fieldName[0] == strings.ToLower(fieldName)[0] {
			continue
		}

		if reflect.DeepEqual(
			resourceField.Interface(),
			reflect.Zero(resourceField.Type()).Interface(),
		) {
			envName := structField.Tag.Get("env")
			if envName != "" {
				envValue := os.Getenv(envName)
				if envValue != "" {
					if !resourceField.CanAddr() {
						return fmt.Errorf(
							"target field is not addressable %q",
							strings.Join(
								push(prefix, getFieldKey(structField)),
								".",
							),
						)
					}

					err := yaml.Unmarshal(
						[]byte(envValue),
						resourceField.Addr().Interface(),
					)
					if err != nil {
						return karma.Format(
							err,
							"unable to unmarshal env value for field: %s",
							strings.Join(
								push(prefix, getFieldKey(structField)),
								".",
							),
						)
					}
				}
			}
		}

		for {
			if resourceField.Kind() != reflect.Ptr {
				break
			}

			if resourceField.IsNil() {
				break
			}

			resourceField = resourceField.Elem()
		}

		if resourceField.Kind() == reflect.Struct && resourceField.CanAddr() {
			err := validate(
				resourceField.Addr().Interface(),
				structFieldRequired,
				push(prefix, getFieldKey(structField))...,
			)
			if err != nil {
				return err
			}
		}

		if reflect.DeepEqual(
			resourceField.Interface(),
			reflect.Zero(resourceField.Type()).Interface(),
		) {
			defaultValue := structField.Tag.Get("default")
			if defaultValue != "" {
				if !resourceField.CanAddr() {
					return fmt.Errorf(
						"target field is not addressable %q",
						strings.Join(
							push(prefix, getFieldKey(structField)),
							".",
						),
					)
				}

				err := yaml.Unmarshal(
					[]byte(defaultValue),
					resourceField.Addr().Interface(),
				)
				if err != nil {
					return karma.Format(
						err,
						"unable to unmarshal default value for field %q",
						strings.Join(
							push(prefix, getFieldKey(structField)),
							".",
						),
					)
				}
			} else if parentRequired && structFieldRequired {
				envName := structField.Tag.Get("env")
				additional := ""
				if envName != "" {
					additional = ", no value for environment variable " +
						envName + " specified"
				}
				return fmt.Errorf(
					"field %q is required, but no value specified%s",
					strings.Join(
						push(prefix, getFieldKey(structField)),
						".",
					),
					additional,
				)
			}
		}

		if resourceField.Kind() == reflect.Slice {
			for i := 0; i < resourceField.Len(); i++ {
				field := reflect.Indirect(resourceField.Index(i))
				if field.Kind() == reflect.Struct {
					err := validate(
						field.Addr().Interface(),
						structFieldRequired,
						push(
							prefix,
							fmt.Sprintf(
								"%s[%d]", getFieldKey(structField), i,
							),
						)...,
					)
					if err != nil {
						return err
					}
				}
			}
		}

		if resourceField.Kind() == reflect.Map {
			for _, key := range resourceField.MapKeys() {
				field := resourceField.MapIndex(key)

				if reflect.Indirect(reflect.ValueOf(field.Interface())).Kind() == reflect.Struct {
					err := validate(
						field.Interface(),
						structFieldRequired,
						push(
							prefix,
							fmt.Sprintf(
								"%s[%s]", getFieldKey(structField), key,
							),
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

func getFieldKey(field reflect.StructField) string {
	knownTags := []string{"yaml", "toml", "json"}
	for _, tag := range knownTags {
		value, ok := field.Tag.Lookup(tag)
		if !ok || value == "" || value == "-" {
			continue
		}

		parts := strings.Split(value, ",")
		return parts[0]
	}

	return strcase.ToSnake(field.Name)
}

func push[K any](prefix []K, value K) []K {
	return append(append([]K{}, prefix...), value)
}
