package common

import (
	"context"
	"io"
	"sync"

	"github.com/calico32/kdl-go"
)

type Hooks struct {
	WhenError   func(mn string, err error)
	WhenLog     func(mn string, f string, msg string)
	WhenSuccess func(mn string, f string, msg string)
}

type Config struct {
	Enabled []string // `kdl:"enabled,args"`
	Paths   map[string]string
}

func ConfigFromKDL(r io.Reader) (*Config, error) {
	doc, err := kdl.Parse(r)
	if err != nil {
		return nil, err
	}
	c := &Config{
		Paths: map[string]string{},
	}
	for _, node := range doc.Nodes {
		name := node.Name()
		switch name {
		case "enabled":
			for _, cNode := range node.Children().Nodes {
				c.Enabled = append(c.Enabled, cNode.Name())
			}
		case "paths":
			for _, cNode := range node.Children().Nodes {
				c.Paths[cNode.Name()] = cNode.Arg(0).String()
			}
		}
	}
	return c, err
}

func (c *Config) ToKDL(w io.Writer) error {
	e := kdl.NewNode("enabled")
	for _, m := range c.Enabled {
		e.AddChild(kdl.NewNode(m))
	}
	p := kdl.NewNode("paths")
	for k, v := range c.Paths {
		node := kdl.NewNode(k)
		node.AddArgument(kdl.NewValue(v))
		p.AddChild(node)
	}
	doc := kdl.NewDocument(e, p)
	return kdl.Encode(doc, w)
}

type ModMeta struct {
	Domain  string   `kdl:"domain"`
	Version string   `kdl:"version"`
	Modules []string `kdl:"modules"`
}

type ModPair struct {
	Meta ModMeta
	Path string
}

type ModEntry struct {
	Pair    ModPair
	Enabled bool
}

type RunData struct {
	Config    *Config
	Context   context.Context
	WaitGroup *sync.WaitGroup
	Hooks     Hooks
	Block     bool
	Here      string
}

type ModuleMeta struct {
	Name      string
	PathNames []string
}

type Module interface {
	Meta() ModuleMeta
	Run(gsEntries []ModEntry, rd RunData) (bool, error)
}
