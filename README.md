# cliche: Simple CLIs for Go

```go
//go:generate TODO(christian)
type Hello struct {
    Name string
}

func (h *Hello) Run(ctx context.Context, out io.Writer) error {
    fmt.Fprintf(out, "Hello, %v!", h.Name)
}
```

After running `go generate` and `go build`, there is now a `hello` command:

```console
$ hello -name=World
Hello, World!
...
```



```go
//go:generate TODO(christian)
type Hello struct {
    Name string `cli:"arg"`
}

func (h *Hello) Run(ctx context.Context, out io.Writer) error {
    fmt.Fprintf(out, "Hello, %v!", h.Name)
}
```

```console
$ hello World
Hello, World!
...
```


```go
//go:generate TODO(christian)
type Hello struct {
    Name string `cli:"default:World"`
}

func (h *Hello) Run(ctx context.Context, out io.Writer) error {
    fmt.Fprintf(out, "Hello, %v!", h.Name)
}
```

```console
$ hello World
Hello, World!
...
```

```go
//go:generate TODO(christian)
type Hello struct {
    Name string `cli:"flag;default:World"`
}

func (h *Hello) Run(ctx context.Context, out io.Writer) error {
    fmt.Fprintf(out, "Hello, %v!", h.Name)
}
```

```console
$ hello World
Hello, World!
...
```
