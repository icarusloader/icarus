// icarus-specific code that does not fit the primary module file
package module

import (
	"encoding/json"
	"errors"
	"fmt"
	"icarus/common"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/KarpelesLab/reflink"
)

type LinkType string

const (
	LinkHard    LinkType = "hard"
	LinkRef     LinkType = "ref"
	LinkCopy    LinkType = "copy"
	LinkDefault LinkType = "default"
)

// map[k]v where k is the file's path and v is the mod's dir
type modMap map[string]string

func build(entries []common.ModEntry) (modMap, error) {
	m := modMap{}
	for _, v := range entries {
		f := os.DirFS(filepath.Join(v.Pair.Path, "basic"))
		if err := fs.WalkDir(f, ".", func(p string, d fs.DirEntry, werr error) error {
			if !d.IsDir() {
				m[p] = v.Pair.Path
			}
			return werr
		}); err != nil {
			return m, err
		}
	}
	return m, nil
}

func applyOverlay(rd common.RunData, m modMap, df string, lt LinkType) error {
	mftPath := filepath.Join(rd.Here, "etc", "basic", "manifest.json")
	arcPath := filepath.Join(rd.Here, "archive", "basic")
	var mft modMap
	curr := modMap{}
	var mftf *os.File
	dirsMade := map[string]struct{}{}
	if _, err := os.Lstat(mftPath); errors.Is(err, fs.ErrNotExist) {
		mftf, err = os.Create(mftPath)
		if err != nil {
			return err
		}
		defer common.CloseInto(&err, mftf)
		mft = modMap{}
		rd.Hooks.WhenLog("basic", "", "created manifest")
	} else if err == nil {
		mftf, err = os.OpenFile(mftPath, os.O_RDWR, 0o644)
		if err != nil {
			return err
		}
		defer common.CloseInto(&err, mftf)
		mftb, err := io.ReadAll(mftf)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(mftb, &mft); err != nil {
			return err
		}
		rd.Hooks.WhenLog("basic", "", "loaded manifest")
	} else {
		return err
	}
	for k, v := range mft {
		dstPath := filepath.Join(df, k)
		arcfPath := filepath.Join(arcPath, k)
		d := filepath.Dir(arcfPath)
		if _, ok := dirsMade[d]; !ok {
			if err := os.MkdirAll(d, 0o755); err != nil {
				return err
			}
			dirsMade[d] = struct{}{}
		}
		if mod, ok := m[k]; mod != v {
			if _, err := os.Lstat(dstPath); err != nil {
				if !errors.Is(err, fs.ErrNotExist) {
					return err
				}
			} else {
				if err := os.Remove(dstPath); err != nil {
					return err
				}
				rd.Hooks.WhenLog("basic", "deleted %s from archive", dstPath)
			}
			if !ok {
				if _, err := os.Lstat(arcfPath); err != nil {
					if !errors.Is(err, fs.ErrNotExist) {
						return err
					}
				} else {
					if err := os.Rename(arcfPath, dstPath); err != nil {
						return err
					}
					rd.Hooks.WhenLog("basic", "restored %s from archive", dstPath)
				}
			}
		}
	}
	for k, v := range m {
		rd.Hooks.WhenLog("basic", "", fmt.Sprintf("working with %s: %s", k, v))
		srcPath := filepath.Join(v, "basic", k)
		dstPath := filepath.Join(df, k)
		arcfPath := filepath.Join(arcPath, k)
		if _, err := os.Lstat(dstPath); err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return err
			}
		} else {
			if _, ok := mft[k]; ok {
				if err := os.Remove(dstPath); err != nil {
					return err
				}
			} else {
				if _, err := os.Lstat(arcfPath); err != nil {
					if !errors.Is(err, fs.ErrNotExist) {
						return err
					}
				} else {
					if err := os.Remove(arcfPath); err != nil {
						return err
					}
					rd.Hooks.WhenLog("basic", "removed %s from archive", dstPath)
				}
				d := filepath.Dir(arcfPath)
				if _, ok := dirsMade[d]; !ok {
					if err := os.MkdirAll(d, 0o755); err != nil {
						return err
					}
					dirsMade[d] = struct{}{}
				}
				if err := os.Rename(dstPath, arcfPath); err != nil {
					return err
				}
				rd.Hooks.WhenLog("basic", "archived %s", dstPath)
			}

		}
		d := filepath.Dir(dstPath)
		if _, ok := dirsMade[d]; !ok {
			if err := os.MkdirAll(d, 0o755); err != nil {
				return err
			}
			dirsMade[d] = struct{}{}
		}
		switch lt {
		case "hard":
			if err := os.Link(srcPath, dstPath); err != nil {
				return err
			}
		case "copy":
			src, err := os.Open(srcPath)
			if err != nil {
				return err
			}
			defer common.CloseInto(&err, src)
			dst, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0o644)
			if err != nil {
				return err
			}
			defer common.CloseInto(&err, dst)
			if _, err = io.Copy(dst, src); err != nil {
				return err
			}
			if err := os.Chmod(dstPath, 0o644); err != nil {
				return err
			}
		case "ref":
			if err := reflink.Always(srcPath, dstPath); err != nil {
				return err
			}
			if err := os.Chmod(dstPath, 0o644); err != nil {
				return err
			}
		default:
			if err := reflink.Always(srcPath, dstPath); err != nil {
				lt = "hard"
				if err := os.Link(srcPath, dstPath); err != nil {
					return err
				}
			} else {
				if err := os.Chmod(dstPath, 0o644); err != nil {
					return err
				}
			}
		}
		rd.Hooks.WhenLog("basic", "wrote to %s", dstPath)
		curr[k] = v
		currb, err := json.Marshal(curr)
		if err != nil {
			return err
		}
		if err = mftf.Truncate(0); err != nil {
			return err
		}
		if _, err = mftf.Seek(0, io.SeekStart); err != nil {
			return err
		}
		if _, err = mftf.Write(currb); err != nil {
			return err
		}
	}

	return nil
}
