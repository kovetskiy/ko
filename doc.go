// Package ko loads configuration files into Go structs with
// struct-tag-driven validation, default values, and environment
// variable fallbacks.
//
// Any format that unmarshals into a struct works (TOML, YAML,
// JSON). The default unmarshaller is [toml.Unmarshal]. Pass a
// different one to [Load].
//
// # Struct tags
//
// Three tags control post-unmarshal behavior:
//
//   - required:"true" — error if the field is still zero after
//     all fallbacks.
//   - default:"value" — set field to value when zero. The value
//     is unmarshalled via yaml.Unmarshal, so complex literals
//     like "[1, 2, 3]" work.
//   - env:"NAME" — read from environment variable NAME when the
//     field is zero after unmarshalling.
//
// Evaluation order: file value → env → default → required check.
// A field with both default and required never triggers the
// required error.
//
// # Required propagation
//
// Required validation recurses into nested structs only when
// the parent struct field is also required. An optional parent
// that is entirely zero-valued skips child validation:
//
//	type Config struct {
//	    DB struct {
//	        URL string `yaml:"url" required:"true"`
//	    } `yaml:"db" required:"true"`  // DB.URL enforced
//
//	    Metrics struct {
//	        Listen string `yaml:"listen" required:"true"`
//	    } `yaml:"metrics"`             // Metrics.Listen not enforced
//	}
//
// # Optional config file
//
// Pass [RequireFile](false) to skip a missing file and apply
// only defaults and environment variables:
//
//	ko.Load(path, &cfg, yaml.Unmarshal, ko.RequireFile(false))
//
// # Pointer fields
//
// Pointer fields distinguish "not set" (nil) from "zero value".
// A *bool with default:"false" allocates a bool pointing to
// false; without the default the pointer stays nil.
//
// # Maps
//
// Map values must be pointers if you need ko to apply defaults
// or env fallbacks to their fields. Non-pointer map values are
// not addressable.
//
// # Field name resolution
//
// Error messages use the yaml, toml, or json struct tag (checked
// in that order) to name fields. When no tag is present, the Go
// field name is converted to snake_case.
package ko
