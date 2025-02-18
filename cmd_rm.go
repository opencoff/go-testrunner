// cmd_mutate.go -- implements the "mutate" command

package testrunner

import (
	"fmt"
	"os"
	"path/filepath"
)

type rmCmd struct{}

func (t *rmCmd) Run(env *TestEnv, args []string) error {
	for i := range args {
		arg := args[i]

		key, vals, err := Split(arg)
		if err != nil {
			return err
		}

		if key != "lhs" && key != "rhs" {
			return fmt.Errorf("mutate: unknown keyword %s", key)
		}

		if len(vals) == 0 {
			return fmt.Errorf("mutate: %s is empty?", key)
		}

		if err = t.rm(key, vals, env); err != nil {
			return fmt.Errorf("mutate: %w", err)
		}
	}
	return nil
}

func (t *rmCmd) rm(key string, vals []string, env *TestEnv) error {
	base := env.TestRoot
	for _, nm := range vals {
		if !filepath.IsAbs(nm) {
			nm = filepath.Join(base, key, nm)
		}

		if exists, err := FileExists(nm); err != nil {
			return err
		} else if !exists {
			return fmt.Errorf("%s: doesn't exist", nm)
		}

		env.Log.Debug("rm %s", nm)

		if err := os.Remove(nm); err != nil {
			return fmt.Errorf("rm %s: %w", nm, err)
		}
	}
	return nil
}

func (t *rmCmd) Name() string {
	return "rm"
}

var _ Cmd = &rmCmd{}

func init() {
	// mutate takes no args
	RegisterCommand(&rmCmd{})
}
