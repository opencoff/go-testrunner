// cmd_symlink.go -- implements the "symlink" command

package testrunner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type symlinkCmd struct {
}

// symlink lhs="newname@oldname newname@oldname" rhs="newname@oldname"
func (t *symlinkCmd) Run(env *TestEnv, args []string) error {
	for i := range args {
		arg := args[i]

		key, vals, err := Split(arg)
		if err != nil {
			return err
		}

		if key != "lhs" && key != "rhs" {
			return fmt.Errorf("symlink: unknown keyword %s", key)
		}

		if len(vals) == 0 {
			return fmt.Errorf("symlink: %s is empty?", key)
		}

		if err = t.symlink(key, vals, env); err != nil {
			return fmt.Errorf("symlink: %w", err)
		}
	}
	return nil
}

func (t *symlinkCmd) symlink(key string, vals []string, env *TestEnv) error {
	base := filepath.Join(env.TestRoot, key)

	for _, nm := range vals {
		i := strings.Index(nm, "@")
		if i < 0 {
			return fmt.Errorf("symlink: %s: incorrect format; exp NEWNAME@OLDNAME", nm)
		}

		newnm := nm[:i]
		oldnm := nm[i+1:]

		env.Log.Debug("symlink (new) %s -> (old) %s", newnm, oldnm)
		if !filepath.IsAbs(oldnm) {
			oldnm = filepath.Join(base, oldnm)
		}

		if !filepath.IsAbs(newnm) {
			newnm = filepath.Join(base, newnm)
		}

		/*
			if exists, err := FileExists(oldnm); err != nil {
				return err
			} else if !exists {
				return fmt.Errorf("%s: doesn't exist", oldnm)
			}
		*/

		// make parent dirs
		dir := filepath.Dir(newnm)
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}

		env.Log.Debug("symlink %s --> %s", oldnm, newnm)
		if err := os.Symlink(oldnm, newnm); err != nil {
			return err
		}
	}
	return nil
}

func (t *symlinkCmd) Name() string {
	return "symlink"
}

var _ Cmd = &symlinkCmd{}

func init() {
	// symlink takes no args
	RegisterCommand(&symlinkCmd{})
}
