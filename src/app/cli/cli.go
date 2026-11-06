package cli

import (
	"context"
	"fmt"
	"icarus/common"
	"icarus/internal/core"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/calico32/kdl-go"

	"github.com/urfave/cli/v3"
)

var hooks = common.Hooks{
	WhenError: func(mn string, err error) {
		fmt.Fprintf(os.Stderr, "\x1b[0;97;41m - \x1b[0m An error has happened in module %s: %v\n", mn, err)
	},
	WhenLog: func(mn string, f string, msg string) {
		if f == "" {
			fmt.Printf("\x1b[0;97;43m ? \x1b[0m Log from %s: %s\n", mn, msg)
		} else {
			fmt.Printf("\x1b[0;97;43m ? \x1b[0m Log from %s: %s\n", mn, fmt.Sprintf(f, msg))
		}
	},
	WhenSuccess: func(mn string, f string, msg string) {
		if f == "" {
			fmt.Fprintf(os.Stderr, "\x1b[0;97;42m + \x1b[0m Message in %s: %s\n", mn, msg)
		} else {
			fmt.Fprintf(os.Stderr, "\x1b[0;97;42m + \x1b[0m Message in %s: %s\n", mn, fmt.Sprintf(f, msg))
		}
	},
}

func Main(args []string, modules []common.Module) {
	ex, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	here := filepath.Dir(ex)
	bg := context.Background()
	cmd := &cli.Command{
		Name:        "icarus",
		Usage:       "Modular modloader for Roblox",
		Description: "A cross-platform and modular modloader written in Go for Roblox",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "modules",
				Aliases:  []string{"m"},
				Required: true,
			},
			&cli.BoolFlag{
				Name:    "block",
				Aliases: []string{"b"},
			},
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			newCtx, cancel := context.WithCancel(ctx)
			defer cancel()
			debug := cmd.Bool("debug")
			if !debug {
				hooks.WhenLog = func(_ string, _ string, _ string) {}
			} else {
				hooks.WhenLog("cli", "", "Hello World!")
			}
			mStr := cmd.String("modules")
			mNames := strings.Split(mStr, ",")
			nMap := map[string]struct{}{}
			for _, m := range modules {
				nMap[m.Meta().Name] = struct{}{}
			}
			block := cmd.Bool("block")
			for _, m := range mNames {
				_, ok := nMap[m]
				if !ok {
					return fmt.Errorf("unknown module %s", m)
				}
				hooks.WhenLog("cli", "Loaded module %s", m)
			}
			confPath := filepath.Join(here, "config.kdl")
			confFile, err := os.Open(confPath)
			if err != nil {
				return err
			}
			conf, err := common.ConfigFromKDL(confFile)
			if err != nil {
				return err
			}
			hooks.WhenLog("cli", "", "config found at config.kdl")
			modsDir := filepath.Join(here, "mods")
			objs, err := os.ReadDir(modsDir)
			if err != nil {
				return err
			}
			avMods := map[string]common.ModPair{}
			for _, obj := range objs {
				if obj.IsDir() {
					modF := filepath.Join(modsDir, obj.Name())
					meta := common.ModMeta{}
					metaFile, err := os.Open(filepath.Join(modF, "meta.kdl"))
					if err != nil {
						return err
					}
					err = kdl.Decode(metaFile, &meta)
					if err != nil {
						return err
					}
					hooks.WhenLog("cli", "mod found with name %s", meta.Domain)
					avMods[meta.Domain] = common.ModPair{
						Meta: meta,
						Path: modF,
					}
				}
			}
			var wg sync.WaitGroup
			ord := &common.OrderedMods{
				Order: []common.ModEntry{},
				Paths: map[string]string{},
			}
			for _, mn := range conf.Enabled {
				mod, ok := avMods[mn]
				if !ok {
					return fmt.Errorf("unknown mod %s found in config", mn)
				}
				ord.AddToTop(common.ModEntry{
					Pair:    mod,
					Enabled: true,
				})
			}
			b, err := core.Run(modules, ord, common.RunData{
				Config:    conf,
				Context:   newCtx,
				WaitGroup: &wg,
				Hooks:     hooks,
				Block:     block,
				Here:      here,
			})
			if err != nil {
				return err
			}
			if block {
				if !b {
					hooks.WhenLog("cli", "", "No modules to block for")
				}
				c := make(chan os.Signal, 1)
				signal.Notify(c, os.Interrupt, syscall.SIGTERM)
				<-c
				cancel()
				wg.Wait()
				os.Exit(0)
			}
			return nil
		},
	}
	err = cmd.Run(bg, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
