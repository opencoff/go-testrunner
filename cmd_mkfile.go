// cmd_mkfile.go -- implements the "tree" command

package testrunner

import (
	"fmt"
	"math/rand/v2"
	"path/filepath"
	"time"

	flag "github.com/opencoff/pflag"
)

type mkfileCmd struct {
}

func (t *mkfileCmd) Name() string {
	return "mkfile"
}

type mkFile struct {
	minsz  SizeValue
	maxsz  SizeValue
	target string
	mkdir  bool
}

// mkfile [-t target] entries...
func (t *mkfileCmd) Run(env *TestEnv, args []string) error {
	mk := &mkFile{
		minsz:  SizeValue(1024),
		maxsz:  SizeValue(8 * 1024),
		target: "",
	}

	fs := flag.NewFlagSet(t.Name(), flag.ExitOnError)
	fs.VarP(&mk.minsz, "min-file-size", "m", "Minimum file size to be created [1k]")
	fs.VarP(&mk.maxsz, "max-file-size", "M", "Maximum file size to be created [8k]")
	fs.BoolVarP(&mk.mkdir, "dir", "d", false, "Make directories instead of files")
	fs.StringVarP(&mk.target, "target", "t", "lhs", "Make entries in the given location (lhs, rhs, both)")
	err := fs.Parse(args)
	if err != nil {
		return fmt.Errorf("mkfile: %w", err)
	}

	env.Log.Debug("mkfile: '%s': sizes: min %d max %d\n", mk.target,
		mk.minsz.Value(), mk.maxsz.Value())

	args = fs.Args()
	now := env.Start
	switch mk.target {
	case "lhs":
		err = mk.mkfile("lhs", args, env, now)
	case "rhs":
		err = mk.mkfile("rhs", args, env, now)
	case "both":
		if err = mk.mkfile("lhs", args, env, now); err != nil {
			return fmt.Errorf("mkfile: %w", err)
		}

		if err = mk.cloneLhs(args, env); err != nil {
			return fmt.Errorf("mkfile: %w", err)
		}
	default:
		return fmt.Errorf("mkfile: unknown target direction '%s'", mk.target)
	}

	if err != nil {
		return fmt.Errorf("mkfile: %w", err)
	}
	return nil
}

func (t *mkFile) cloneLhs(args []string, env *TestEnv) error {
	base := env.TestRoot
	for _, nm := range args {
		if filepath.IsAbs(nm) {
			return fmt.Errorf("common file %s can't be absolute", nm)
		}

		lhs := filepath.Join(base, "lhs", nm)
		rhs := filepath.Join(base, "rhs", nm)
		env.Log.Debug("mkfile clone %s -> %s", lhs, rhs)
		if err := copyfile(rhs, lhs); err != nil {
			return fmt.Errorf("%s: %w", rhs, err)
		}
	}
	return nil
}

func (t *mkFile) mkfile(key string, args []string, env *TestEnv, now time.Time) error {
	base := env.TestRoot
	for _, nm := range args {
		var err error
		fn := nm

		if !filepath.IsAbs(nm) {
			fn = filepath.Join(base, key, fn)
		}

		if t.mkdir {
			env.Log.Debug("mkdir %s", fn)
			err = mkdir(fn, now)
		} else {
			sz := int64(rand.N(t.maxsz-t.minsz) + t.minsz)
			env.Log.Debug("mkfile %s %d", fn, sz)
			err = mkfile(fn, sz, now)
		}

		if err != nil {
			return fmt.Errorf("%s: %w", fn, err)
		}
	}
	return nil
}

var _ Cmd = &mkfileCmd{}

func init() {
	RegisterCommand(&mkfileCmd{})
}
