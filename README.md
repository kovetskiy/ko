# ko

ko loads configuration files into Go structs. Any format that
can unmarshal into a struct works: TOML (default), YAML, JSON.

Struct tags control validation and fallback values. ko checks
them after unmarshalling, so the file format is irrelevant to
the tag behavior.

## Tags

Three tags, applied to struct fields:

| Tag | Value | Effect |
|-----|-------|--------|
| `required` | `"true"` | Error if field is zero after load |
| `default` | any string | Set field to this value if zero |
| `env` | env var name | Read from environment if zero |

Evaluation order: file value → environment variable → default
→ required check. A field with both `default` and `required`
never fails the required check because the default fills it.

Default values are unmarshalled through `yaml.Unmarshal`, so
complex types work: `default:"[1, 2, 3]"` fills an `[]int`.

## Basic usage

```go
type Config struct {
    Listen   string `yaml:"listen"   default:":8080"`
    Database string `yaml:"database" required:"true" env:"DB_URL"`
    Debug    bool   `yaml:"debug"    default:"false" env:"DEBUG"`
}

var cfg Config
err := ko.Load("config.yaml", &cfg, yaml.Unmarshal)
```

If `config.yaml` sets `database`, that value wins. If not, ko
checks `$DB_URL`. If the env var is also empty, Load returns
an error because the field is required.

## Custom unmarshaller

The default unmarshaller is `toml.Unmarshal`. Pass a different
one as a variadic argument:

```go
ko.Load("config.yaml", &cfg, yaml.Unmarshal)
ko.Load("config.json", &cfg, json.Unmarshal)
```

## Optional config file

By default, Load fails if the file does not exist. Pass
`ko.RequireFile(false)` to skip missing files and rely
entirely on defaults and environment variables:

```go
err := ko.Load(path, &cfg, yaml.Unmarshal, ko.RequireFile(false))
```

## Nested structs and required propagation

Required validation propagates into child structs only when
the parent is also marked `required:"true"`.

```go
type Config struct {
    Server struct {
        Host string `yaml:"host" required:"true"`
        Port int    `yaml:"port" required:"true" default:"443"`
    } `yaml:"server" required:"true"`

    Metrics struct {
        Listen string `yaml:"listen" required:"true" default:":9090"`
    } `yaml:"metrics"`
}
```

`Server.Host` must be set because both `Server` and `Host`
are required. `Metrics.Listen` is never enforced even though
the field says `required:"true"`, because `Metrics` itself is
not required. When `Metrics` is zero-valued (nothing in the
file), ko skips validation of its children entirely.

This lets you define optional sections that, once partially
filled, still enforce their own constraints.

## Slices and maps

ko validates elements inside slices and maps when the parent
field is required:

```go
type Route struct {
    Path    string `yaml:"path"    required:"true"`
    Backend string `yaml:"backend" required:"true"`
}

type Config struct {
    Routes []Route `yaml:"routes" required:"true"`
}
```

For maps, use pointer values if you need defaults or env
fallbacks applied to map entries:

```go
type Config struct {
    Workers map[string]*WorkerConfig `yaml:"workers" required:"true"`
}
```

Non-pointer map values are not addressable, so ko cannot set
defaults or env values on their fields. It returns an error
like `target field is not addressable "workers[key].field"`.

## Pointer fields

Pointer fields distinguish "not set" from "zero value":

```go
type Config struct {
    Verbose *bool `yaml:"verbose" default:"false"`
}
```

When the file omits `verbose`, ko allocates a `*bool` pointing
to `false`. Without the default, the pointer stays `nil`. This
matters for boolean flags where `false` is a meaningful value
different from "not configured".

## Field name resolution

ko uses struct tags to build field paths for error messages.
It checks `yaml`, `toml`, and `json` tags in that order. If
none are present, it converts the Go field name to snake_case.

## Kogen

kogen generates a documentation table from a config struct:

```
go run ./cmd/kogen /path/to/source StructName
```

Output is a markdown table with columns: Variable, Environment
Variable, Default Value, Type, Required.
