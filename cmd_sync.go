// cmd_sync.go -- implements the "sync" command

package testrunner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type syncCmd struct {
}

func (t *syncCmd) Run(env *TestEnv, args []string) error {
	dirs := []string{
		env.Lhs,
		env.Rhs,
	}

	// first adjtime for all non-dir entries
	now := env.Start
	for _, dn := range dirs {
		err := filepath.Walk(dn, func(p string, fi fs.FileInfo, err error) error {
			if fi.IsDir() {
				return nil
			}
			if fi.Mode().Type() != fs.ModeSymlink {
				err := os.Chtimes(p, now, now)
				if err != nil {
					return fmt.Errorf("adjtime: %s %w", p, err)
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	// now do the same for dirs
	for _, dn := range dirs {
		err := filepath.Walk(dn, func(p string, fi fs.FileInfo, err error) error {
			if fi.IsDir() {
				err := os.Chtimes(p, now, now)
				if err != nil {
					return fmt.Errorf("adjtime: %s %w", p, err)
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *syncCmd) Name() string {
	return "sync"
}

var _ Cmd = &syncCmd{}

func init() {
	RegisterCommand(&syncCmd{})
}
