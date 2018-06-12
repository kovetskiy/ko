package ko

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
		`B is required, but no value specified`,
	)

	resource.B.X = true
	test.EqualError(
		Load(path, &resource),
		`B.Y is required, but no value specified`,
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
		`B is required, but no value specified`,
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
		`C is required, but no value specified`,
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
