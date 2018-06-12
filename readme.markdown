# ko [![GoDoc](https://godoc.org/github.com/kovetskiy/ko?status.svg)](http://godoc.org/github.com/kovetskiy/ko) [![Go Report Card](https://goreportcard.com/badge/github.com/kovetskiy/ko)](https://goreportcard.com/report/github.com/kovetskiy/ko) [![Build Status](https://travis-ci.org/kovetskiy/ko.svg?branch=master)](https://travis-ci.org/kovetskiy/ko)

ko is a package for golang that allows to use configuration files with any format
that can be unmarshalled to Go struct like TOML/YAML/JSON.

## Supported tags

ko is controlled by Go struct tags, following tags are supported:
- `required` - if specified as true, will trigger an error if no value
    specified for field.
- `env` - read value from specified environment variable if no value specified
    in config file.
- `default` - specifies default value for field, specified value will be
    unmarshalled using yaml.Unmarshal.

## Usage

```go
type Config struct {
    Hostname string `default:"localhost"`
    Username string `required:"true"`
    Port     int    `env:"PORT" default:"8086"`
    Password string
}

var config Config
err := ko.Load("app.conf", &config)
if err != nil  {
    panic(err)
}
```

Will return error if no value for Username field specified, Hostname will be
`localhost` if no value specified in config file, Password will be empty if no
value specified, no error will be returned for this field.
If PORT environment variable is specified, Port field will contain its value,
if given environment variable doesn't exist, default value `8086` will be used.

Also, you can pass custom unmarshaller if you use custom format.

```go
type Config struct {
    Hostname string `default:"localhost"`
    Username string `required:"true"`
    Password string
}

var config Config
err := ko.Load("app.json", &config, json.Unmarshal)
if err != nil {
    panic(err)
}
```


The package doesn't support maps.

## License

MIT.
