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
	Flags   map[string]map[string]any
}

func ConfigFromKDL(r io.Reader) (*Config, error) {
	doc, err := kdl.Parse(r)
	if err != nil {
		return nil, err
	}
	c := &Config{
		Flags: map[string]map[string]any{},
	}
	for _, node := range doc.Nodes {
		name := node.Name()
		switch name {
		case "enabled":
			for _, cNode := range node.Children().Nodes {
				c.Enabled = append(c.Enabled, cNode.Name())
			}
		case "flags":
			for _, cNode := range node.Children().Nodes {
				for _, ccNode := range cNode.Children().Nodes {
					if ccNode.Arg(0).Kind() == kdl.Bool {
						c.Flags[cNode.Name()][ccNode.Name()] = ccNode.Arg(0).Bool()
					} else if ccNode.Arg(0).Kind() == kdl.String {
						c.Flags[cNode.Name()][ccNode.Name()] = ccNode.Arg(0).String()
					}
				}
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
	f := kdl.NewNode("flags")
	for k, v := range c.Flags {
		pNode := kdl.NewNode(k)
		for ck, cv := range v {
			node := kdl.NewNode(ck)
			val, err := kdl.TryNewValue(cv)
			if err != nil {
				return err
			}
			node.AddArgument(val)
			pNode.AddChild(node)
		}
	}
	doc := kdl.NewDocument(e, f)
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
	Name  string
	Flags []string
}

type Module interface {
	Meta() ModuleMeta
	Run(gsEntries []ModEntry, rd RunData) (bool, error)
}
