# ko [![GoDoc](https://godoc.org/github.com/kovetskiy/ko?status.svg)](http://godoc.org/github.com/kovetskiy/ko) [![Go Report Card](https://goreportcard.com/badge/github.com/kovetskiy/ko)](https://goreportcard.com/report/github.com/kovetskiy/ko) [![Build Status](https://travis-ci.org/kovetskiy/ko.svg?branch=master)](https://travis-ci.org/kovetskiy/ko)


ko is a package for golang that allow to use configuration files with any format
that can be unmarshalled to golang structures like TOML/YAML/JSON.

## Usage

```go
type Config struct {
    Hostname string `default:"localhost"`
    Username string `required:"true"`
    Password string
}

var config Config
ko.Load("app.conf", &config)
```

## License

MIT.
