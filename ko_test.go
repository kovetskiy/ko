package ko

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestUnmarshall(t *testing.T) {
	test := assert.New(t)

	path := write(`
a = true
[b]
c = "c"
d = 3
[[e]]
x = "www"
`)
	defer os.Remove(path)

	type configSliceField struct {
		X string
	}

	type config struct {
		A bool
		B struct {
			C string
			D int64
		}
		E []configSliceField
	}

	resource := config{}
	test.NoError(
		Load(path, &resource),
	)

	ko := config{}
	ko.A = true
	ko.B.C = "c"
	ko.B.D = 3
	ko.E = []configSliceField{
		{X: "www"},
	}

	test.Equal(ko, resource)
}

func TestUnmarshal_IntoSlice(t *testing.T) {
	test := assert.New(t)

	path := write(`
- a: 1
  b: 2
- a: 3
`)
	defer os.Remove(path)

	type config struct {
		A int
		B int
	}

	var result []config
	test.NoError(
		Load(path, &result, yaml.Unmarshal),
	)

	test.Len(result, 2)
	test.Equal(config{A: 1, B: 2}, result[0])
	test.Equal(config{A: 3}, result[1])
}

func TestDefault(t *testing.T) {
	test := assert.New(t)

	path := write(`
b = true
`)
	defer os.Remove(path)

	type config struct {
		A string `default:"aaa"`
	}

	resource := config{}
	test.NoError(
		Load(path, &resource),
	)

	test.Equal("aaa", resource.A)
}

func TestEnv(t *testing.T) {
	test := assert.New(t)

	path := write(`
b = true
`)
	defer os.Remove(path)

	type config struct {
		A string `env:"aaa"`
	}

	os.Setenv("aaa", "valueA")
	defer os.Setenv("aaa", "")

	resource := config{}
	test.NoError(
		Load(path, &resource),
	)

	test.Equal("valueA", resource.A)
}

func TestEnvMissing(t *testing.T) {
	test := assert.New(t)

	path := write(`
b = true
`)
	defer os.Remove(path)

	type config struct {
		A string `env:"bbb"`
	}

	resource := config{}
	test.NoError(
		Load(path, &resource),
	)

	test.Equal("", resource.A)
}

func TestEnvDefault(t *testing.T) {
	test := assert.New(t)

	path := write(`
b = true
`)
	defer os.Remove(path)

	type config struct {
		A string `env:"bbb" default:"123"`
	}

	resource := config{}
	test.NoError(
		Load(path, &resource),
	)

	test.Equal("123", resource.A)
}

func TestEnvRequired(t *testing.T) {
	test := assert.New(t)

	path := write(`
b = true
`)
	defer os.Remove(path)

	type config struct {
		A string `env:"bbb" required:"true"`
	}

	resource := config{}
	test.Error(
		Load(path, &resource),
	)
}

func TestRequired(t *testing.T) {
	test := assert.New(t)

	path := write(``)
	defer os.Remove(path)

	type config struct {
		A string `required:"true"`
	}

	resource := config{}
	test.Error(
		Load(path, &resource),
	)
}

func TestRequiredInSubFieldButNoInParent(t *testing.T) {
	test := assert.New(t)

	path := write(``)
	defer os.Remove(path)

	type config struct {
		A bool
		B struct {
			X bool
			Y bool `required:"true"`
		}
	}

	resource := config{}
	test.NoError(
		Load(path, &resource),
	)
}

func TestRequiredInSubFieldAndParentRequired(t *testing.T) {
	test := assert.New(t)

	path := write(``)
	defer os.Remove(path)

	type config struct {
		A bool
		B struct {
			X bool
			Y bool `required:"true"`
		} `required:"true"`
	}

	resource := config{}
	test.EqualError(
		Load(path, &resource),
		`field "b.y" is required, but no value specified`,
	)

	resource.B.X = true
	test.EqualError(
		Load(path, &resource),
		`field "b.y" is required, but no value specified`,
	)
}

func TestRequiredStructSubField(t *testing.T) {
	test := assert.New(t)

	path := write(``)
	defer os.Remove(path)

	type config struct {
		A bool
		B struct {
			X bool
			Y bool
		} `required:"true"`
	}

	resource := config{}
	test.EqualError(
		Load(path, &resource),
		`field "b" is required, but no value specified`,
	)
}

func TestRequiredAfterStruct(t *testing.T) {
	test := assert.New(t)

	path := write(``)
	defer os.Remove(path)

	type config struct {
		A bool
		B struct {
			X bool
			Y bool
		}
		C bool `required:"true"`
	}

	resource := config{}
	test.EqualError(
		Load(path, &resource),
		`field "c" is required, but no value specified`,
	)
}

func TestRequiredAndDefault(t *testing.T) {
	test := assert.New(t)

	path := write(`
b = true
`)
	defer os.Remove(path)

	type config struct {
		A string `required:"true" default:"aaa"`
	}

	resource := config{}
	test.NoError(
		Load(path, &resource),
	)

	test.Equal("aaa", resource.A)
}

func TestRequiredUseYamlTag(t *testing.T) {
	test := assert.New(t)

	path := write(``)
	defer os.Remove(path)

	type config struct {
		A string `yaml:"blah,omitempty" required:"true"`
	}

	resource := config{}
	test.EqualError(
		Load(path, &resource),
		`field "blah" is required, but no value specified`,
	)
}

func TestRequiredUseTomlTag(t *testing.T) {
	test := assert.New(t)

	path := write(``)
	defer os.Remove(path)

	type config struct {
		A string `toml:"blah,omitempty" required:"true"`
	}

	resource := config{}
	test.EqualError(
		Load(path, &resource),
		`field "blah" is required, but no value specified`,
	)
}

func TestRequiredUseJsonTag(t *testing.T) {
	test := assert.New(t)

	path := write(``)
	defer os.Remove(path)

	type config struct {
		A string `json:"blah" required:"true"`
	}

	resource := config{}
	test.EqualError(
		Load(path, &resource),
		`field "blah" is required, but no value specified`,
	)
}

func TestPointerField_ValueNotSpecified(t *testing.T) {
	test := assert.New(t)

	path := write(``)
	defer os.Remove(path)

	type config struct {
		A *bool `json:"a"`
	}

	resource := config{}
	test.NoError(Load(path, &resource, yaml.Unmarshal))
	test.Nil(resource.A)
}

func TestPointerField_ValueSpecified(t *testing.T) {
	test := assert.New(t)

	path := write(`a: false`)
	defer os.Remove(path)

	type config struct {
		A *bool `json:"a"`
	}

	resource := config{}
	test.NoError(Load(path, &resource, yaml.Unmarshal))
	test.NotNil(resource.A)
	test.False(*resource.A)
}

func TestPointerField_ValueNotSpecified_UseDefault(t *testing.T) {
	test := assert.New(t)

	path := write(``)
	defer os.Remove(path)

	type config struct {
		A *bool `json:"a" default:"false"`
	}

	resource := config{}
	test.NoError(Load(path, &resource, yaml.Unmarshal))
	if test.NotNil(resource.A) {
		test.False(*resource.A)
	}
}

func TestPointerFields_ValueNotSpecified(t *testing.T) {
	test := assert.New(t)

	path := write(``)
	defer os.Remove(path)

	type Blah struct {
		A *bool `json:"a"`
	}

	type config struct {
		Blah *Blah `json:"blah"`
	}

	resource := config{}
	test.NoError(Load(path, &resource, yaml.Unmarshal))
	test.Nil(resource.Blah)
}

func TestPointerFields_ValueSpecified(t *testing.T) {
	test := assert.New(t)

	path := write(`
blah:
    a: true
`)
	defer os.Remove(path)

	type Blah struct {
		A *bool `json:"a"`
	}

	type config struct {
		Blah *Blah `json:"blah" required:"true"`
	}

	resource := config{}
	test.NoError(Load(path, &resource, yaml.Unmarshal))
	test.NotNil(resource.Blah)
	test.NotNil(resource.Blah.A)
	test.True(*resource.Blah.A)
}

func TestSkipUnexportedFields(t *testing.T) {
	test := assert.New(t)

	path := write(`
a = true
`)
	defer os.Remove(path)

	type config struct {
		A bool

		unexported bool
	}

	resource := config{}
	test.NoError(
		Load(path, &resource),
	)

	ko := config{}
	ko.A = true
	ko.unexported = false

	test.Equal(ko, resource)
}

func write(data string) string {
	file, err := ioutil.TempFile(os.TempDir(), "ko_")
	if err != nil {
		panic(err)
	}

	_, err = file.WriteString(data)
	if err != nil {
		panic(err)
	}

	return file.Name()
}

func TestRequireFileOpt(t *testing.T) {
	test := assert.New(t)

	path := write(`
sequence = "AATGAGTC"
`)

	defer os.Remove(path)

	type config struct {
		Sequence string `default:"UNKNOWN"`
	}

	{
		var resource config
		test.NoError(Load(path, &resource, RequireFile(false)))
		test.Equal("AATGAGTC", resource.Sequence)
	}

	test.NoError(os.Remove(path))

	{
		var resource config
		test.NoError(Load(path, &resource, RequireFile(false)))
		test.Equal("UNKNOWN", resource.Sequence)
	}

	{
		var resource config
		test.Error(Load(path, &resource, RequireFile(true)))
	}
}

func TestDoNotError_ForRequiredButFilledWithDefault(t *testing.T) {
	test := assert.New(t)

	type resource struct {
		Foo string `yaml:"foo" required:"true" env:"FOO" default:"default-foo"`
		Bar string `yaml:"bar" required:"true" env:"BAR" default:"bar"`
	}

	type config struct {
		Resource resource `required:"true"`
	}

	{
		var cfg config
		err := Load("/does/not/exist", &cfg, RequireFile(false))
		test.NoError(err)
	}
}

func TestMeaningfulErrorForRequiredStruct(t *testing.T) {
	test := assert.New(t)

	type resource struct {
		Foo string `yaml:"foo" required:"true" env:"FOO" default:"default-foo"`
		Bar string `yaml:"bar" required:"true" env:"BAR" default:""`
	}

	type config struct {
		Resource resource `required:"true"`
	}

	{
		var cfg config
		err := Load("/does/not/exist", &cfg, RequireFile(false))
		test.EqualError(
			err,
			`field "resource.bar" is required, but no value specified, no value for environment variable BAR specified`,
		)
	}
}

func TestCheckRequiredFieldsInMap_Missing(t *testing.T) {
	test := assert.New(t)

	path := write(`
foo:
 key:
   unused: 123
`)
	defer os.Remove(path)

	type data struct {
		Bar string `required:"true"`
	}

	type config struct {
		Foo map[string]data `required:"true"`
	}

	{
		var cfg config
		err := Load(path, &cfg, yaml.Unmarshal)
		test.EqualError(err, `field "foo[key].bar" is required, but no value specified`)
	}
}

func TestCheckRequiredFieldsInMap_Default(t *testing.T) {
	test := assert.New(t)

	path := write(`
foo:
 key:
   unused: 123
`)
	defer os.Remove(path)

	type data struct {
		Bar string `required:"true" default:"q"`
	}

	type config struct {
		Foo map[string]*data `required:"true"`
	}

	{
		var cfg config
		err := Load(path, &cfg, yaml.Unmarshal)
		test.Equal("q", cfg.Foo["key"].Bar)
		test.NoError(err)
	}
}

func TestCheckRequired_DefaultInSubfield(t *testing.T) {
	test := assert.New(t)

	type Bar struct {
		Value string `yaml:"value" required:"true" env:"VALUE" default:"bar-value"`
	}

	type config struct {
		Foo string `yaml:"foo" required:"true" env:"FOO" default:"foo"`
		Bar Bar    `yaml:"bar" required:"true"           default:""`
	}

	{
		var cfg config
		err := Load("", &cfg, yaml.Unmarshal, RequireFile(false))
		test.Equal("foo", cfg.Foo)
		test.Equal("bar-value", cfg.Bar.Value)
		test.NoError(err)
	}
}

func TestCheckRequiredFieldsInMap_DefaultNotaddressable(t *testing.T) {
	test := assert.New(t)

	path := write(`
foo:
 key:
   unused: 123
`)
	defer os.Remove(path)

	type data struct {
		Bar string `required:"true" default:"q"`
	}

	type config struct {
		Foo map[string]data `required:"true"`
	}

	{
		var cfg config
		err := Load(path, &cfg, yaml.Unmarshal)
		test.EqualError(err, `target field is not addressable "foo[key].bar"`)
	}
}

func TestCheckRequiredFieldsInMap_Env(t *testing.T) {
	test := assert.New(t)

	path := write(`
foo:
 key:
   unused: 123
`)
	defer os.Remove(path)

	type data struct {
		Bar string `required:"true" env:"aaa"`
	}

	type config struct {
		Foo map[string]*data `required:"true"`
	}

	os.Setenv("aaa", "valueA")
	defer os.Setenv("aaa", "")

	{
		var cfg config
		err := Load(path, &cfg, yaml.Unmarshal)
		test.Equal("valueA", cfg.Foo["key"].Bar)
		test.NoError(err)
	}
}

func TestCheckRequiredFieldsInMap_EnvNotaddressable(t *testing.T) {
	test := assert.New(t)

	path := write(`
foo:
 key:
   unused: 123
`)
	defer os.Remove(path)

	type data struct {
		Bar string `required:"true" env:"aaa"`
	}

	type config struct {
		Foo map[string]data `required:"true"`
	}

	os.Setenv("aaa", "valueA")
	defer os.Setenv("aaa", "")

	{
		var cfg config
		err := Load(path, &cfg, yaml.Unmarshal)
		test.EqualError(err, `target field is not addressable "foo[key].bar"`)
	}
}

func TestCheckRequiredFieldsInMap_Value(t *testing.T) {
	test := assert.New(t)

	path := write(`
foo:
 key:
   bar: 123
`)
	defer os.Remove(path)

	type data struct {
		Bar string `required:"true"`
	}

	type config struct {
		Foo map[string]data `required:"true"`
	}

	{
		var cfg config
		err := Load(path, &cfg, yaml.Unmarshal)
		test.Equal("123", cfg.Foo["key"].Bar)
		test.NoError(err)
	}
}

func TestCheckMapTypeInt(t *testing.T) {
	test := assert.New(t)

	path := write(`{
	"foo": {
		"": 1,
		"A": 2
	}
}
`)
	defer os.Remove(path)

	type config struct {
		Foo map[string]int64 `required:"true"`
	}

	{
		var cfg config
		err := Load(path, &cfg, json.Unmarshal)
		test.Equal(int64(1), cfg.Foo[""])
		test.Equal(int64(2), cfg.Foo["A"])
		test.NoError(err)
	}
}

func TestOptionalStructWithRequiredFields(t *testing.T) {
	// Test case: Optional struct containing required fields should not
	// validate required fields when the optional struct is not provided

	type RequiredCredentials struct {
		APIKey string `yaml:"api_key" required:"true"`
		Secret string `yaml:"secret" required:"true"`
	}

	type OptionalAuth struct {
		Credentials RequiredCredentials `yaml:"credentials" required:"true"`
		Timeout     int                 `yaml:"timeout" required:"true"`
	}

	type Config struct {
		// This field is optional
		Auth OptionalAuth `yaml:"auth" required:"false"`
		// This field is required
		Port int `yaml:"port" required:"true"`
	}

	// Test 1: Empty config should fail (missing required port)
	t.Run("empty config fails on required field", func(t *testing.T) {
		test := assert.New(t)
		path := write(``)
		defer os.Remove(path)

		var cfg Config
		err := Load(path, &cfg, yaml.Unmarshal)
		test.Error(err)
		test.Contains(err.Error(), "port")
	})

	// Test 2: Config with only port should succeed (auth is optional)
	t.Run("config without optional struct succeeds", func(t *testing.T) {
		test := assert.New(t)
		// testdata/port-only.yaml contains:
		// port: 8080
		path := write(`port: 8080`)
		defer os.Remove(path)

		var cfg Config
		err := Load(path, &cfg, yaml.Unmarshal)
		test.NoError(err) // This currently FAILS but should PASS
		test.Equal(8080, cfg.Port)
	})

	// Test 3: Config with auth but missing required fields should fail
	t.Run("config with incomplete optional struct fails", func(t *testing.T) {
		test := assert.New(t)
		// testdata/incomplete-auth.yaml contains:
		// port: 8080
		// auth:
		//   timeout: 30
		path := write(`port: 8080
auth:
  timeout: 30`)
		defer os.Remove(path)

		var cfg Config
		err := Load(path, &cfg, yaml.Unmarshal)
		test.Error(err)
		test.Contains(err.Error(), "auth.credentials")
	})

	// Test 4: Complete config should succeed
	t.Run("complete config succeeds", func(t *testing.T) {
		test := assert.New(t)
		// testdata/complete.yaml contains:
		// port: 8080
		// auth:
		//   credentials:
		//     api_key: "key123"
		//     secret: "secret456"
		//   timeout: 30
		path := write(`port: 8080
auth:
  credentials:
    api_key: "key123"
    secret: "secret456"
  timeout: 30`)
		defer os.Remove(path)

		var cfg Config
		err := Load(path, &cfg, yaml.Unmarshal)
		test.NoError(err)
		test.Equal(8080, cfg.Port)
		test.Equal("key123", cfg.Auth.Credentials.APIKey)
	})
}
