# ko [![GoDoc](https://godoc.org/github.com/kovetskiy/ko?status.svg)](http://godoc.org/github.com/kovetskiy/ko)

ko is a package for golang that allow to use configuration files with any format
that can be unmarshalled to golang structures like TOML/YAML/JSON.

## Usage

```
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
