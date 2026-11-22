package cli

import (
	"context"
	"errors"
	"fmt"
	"icarus/common"
	"icarus/internal/core"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"syscall"

	"github.com/calico32/kdl-go"

	"github.com/urfave/cli/v3"
)

var hooks = common.Hooks{
	WhenError: func(mn string, err error) {
		fmt.Fprintf(os.Stderr, "\x1b[0;90;41m - \x1b[0m An error has happened in module %s: %v\n", mn, err)
	},
	WhenLog: func(mn string, f string, msg string) {
		if f == "" {
			fmt.Printf("\x1b[0;90;43m ? \x1b[0m Log from %s: %s\n", mn, msg)
		} else {
			fmt.Printf("\x1b[0;90;43m ? \x1b[0m Log from %s: %s\n", mn, fmt.Sprintf(f, msg))
		}
	},
	WhenSuccess: func(mn string, f string, msg string) {
		if f == "" {
			fmt.Fprintf(os.Stderr, "\x1b[0;90;42m + \x1b[0m Message in %s: %s\n", mn, msg)
		} else {
			fmt.Fprintf(os.Stderr, "\x1b[0;90;42m + \x1b[0m Message in %s: %s\n", mn, fmt.Sprintf(f, msg))
		}
	},
	WhenOther: func(mn string, f string, msg string) {
		if f == "" {
			fmt.Fprintf(os.Stderr, "\x1b[0;90;44m ! \x1b[0m Message in %s: %s\n", mn, msg)
		} else {
			fmt.Fprintf(os.Stderr, "\x1b[0;90;44m ! \x1b[0m Message in %s: %s\n", mn, fmt.Sprintf(f, msg))
		}
	},
}

func Main(args []string, modules []common.Module) {
	ex, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "\x1b[0;90;41m - \x1b[0m Error: %v\n", err)
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
				hooks.WhenLog("cli", "loaded module %s", m)
			}
			var conf *common.Config
			confPath := filepath.Join(here, "config.kdl")
			confFile, err := os.Open(confPath)
			defer common.CloseInto(&err, confFile)
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					confFile, err = os.Create(confPath)
					hooks.WhenOther("cli", "", "created config.kdl")
					defer common.CloseInto(&err, confFile)
					if err != nil {
						return err
					}
					conf = &common.Config{
						Enabled: []string{},
						Flags:   map[string]map[string]any{},
					}
					if err = conf.ToKDL(confFile); err != nil {
						return err
					}
				} else {
					return err
				}
			} else {
				conf, err = common.ConfigFromKDL(confFile)
				if err != nil {
					return err
				}
			}
			enabled := make([]common.Module, 0, len(modules))
			for _, m := range modules {
				if slices.Contains(mNames, m.Meta().Name) {
					enabled = append(enabled, m)
				}
			}
			hooks.WhenLog("cli", "", "config found at config.kdl")
			modsDir := filepath.Join(here, "mods")
			objs, err := os.ReadDir(modsDir)
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					if err = os.Mkdir(modsDir, 0o755); err != nil {
						return err
					}
					hooks.WhenOther("cli", "", "created the mods folder, nothing else to do")
					return nil
				} else {
					return err
				}
			}
			avMods := map[string]common.ModPair{}
			for _, obj := range objs {
				if obj.IsDir() {
					modF := filepath.Join(modsDir, obj.Name())
					meta := common.ModMeta{}
					metaFile, err := os.Open(filepath.Join(modF, "meta.kdl"))
					defer common.CloseInto(&err, metaFile)
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
			b, err := core.Run(enabled, ord, common.RunData{
				Config:    conf,
				Context:   newCtx,
				WaitGroup: &wg,
				Hooks:     hooks,
				Block:     block,
				Here:      here,
			})
			if err != nil {
				return nil
			}
			hooks.WhenSuccess("cli", "", fmt.Sprintf("successfully applied %d mod(s)", len(ord.Order)))
			if block {
				if !b {
					hooks.WhenOther("cli", "", "no modules to block for")
					return nil
				}
				hooks.WhenSuccess("cli", "", "blocking...")
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
		hooks.WhenError("cli", err) // i was wrong
		os.Exit(1)
	}
}
