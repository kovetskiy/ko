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

## Kogen

kogen parses source code using ast standard packages, finds specified struct,
and generates documentation based on struct field tags.

```
kogen /path/to/source/code NameOfStruct
```

will find NameOfStruct in /path/to/source/code and generate the following table:

Variable | Environment Variable | Default Value | Type | Required |
--- | --- | --- | --- | --- |
`live.address` | `LIVE_FEED_ADDRESS` | `<no value>` | `string` | true |
`live.buffer_size` | `LIVE_FEED_BUFFER_SIZE` | `1` | `int` | true |
`cache.address` | `CACHE_ADDRESS` | `<no value>` | `string` | true |
`cache.max_call_recv_msg_size` | `CACHE_MAX_CALL_RECV_MSG_SIZE` | `33554432` | `int` | true |
`http.listen` | `HTTP_LISTEN` | `:80` | `string` | true |

The package doesn't support maps.

## License

MIT.
