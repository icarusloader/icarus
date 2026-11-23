# ![](https://wsrv.nl?url=raw.githubusercontent.com/icarusloader/.github/416d97ada95d88f771b4542b41769d8241d99bad/media/icarus.svg&height=32) icarusloader modules API

icarusloader defines a simple API for modules.

```go
type Module struct{}

func (Module) Meta() common.ModuleMeta {
	return common.ModuleMeta{
		Name: "example",
	}
}

func (Module) Run(entries []common.ModEntry, rd common.RunData) (bool, error) {
	block := false // change it depending of your code will block or not
	// do stuff ...
	return block, nil
}
```

\[\]ModEntry provides the mods that your module will use. RunData has the following information that might be helpful:

```go
type RunData struct {
	Config    *Config // the data inside config.kdl
	Context   context.Context // a context
	WaitGroup *sync.WaitGroup // a waitgroup
	Hooks     Hooks // hooks you can use to print and return errors
	Block     bool // whether blocking is enabled or not
	Here      string // the binary's folder
}
```

For more information on this, read the source code of common/primary.go and common/other.go.
