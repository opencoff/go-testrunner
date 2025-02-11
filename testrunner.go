// tmain.go - main test runner

package testrunner

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/opencoff/go-logger"
)

var Z = filepath.Base(os.Args[0])

type Config struct {
	Tempdir string

	LogStdout bool
	Ncpu      int
}

type TestRunner struct {
	Config
}

func New(c *Config) *TestRunner {
	tmp := c.Tempdir
	if len(tmp) == 0 {
		tmp = os.TempDir()
	}

	tr := &TestRunner{
		Config: *c,
	}

	tr.Tempdir = filepath.Join(tmp, fmt.Sprintf("testrunner-%s", randstr(4)))

	if tr.Ncpu == 0 {
		tr.Ncpu = runtime.NumCPU() / 2 // don't use all available cpus
	}
	return tr
}

func (tr *TestRunner) Run(args []string, par bool) error {
	var err error
	if par {
		err = tr.parallelize(args)
	} else {
		err = tr.serialize(args)
	}

	if err == nil {
		err = os.RemoveAll(tr.Tempdir)
	}
	return err
}

func (tr *TestRunner) RunOne(fn string) error {
	var err error

	if err = tr.runTest(fn); err == nil {
		err = os.RemoveAll(tr.Tempdir)
	}
	return err
}

func (tr *TestRunner) serialize(args []string) error {
	for _, fn := range args {
		if err := tr.runTest(fn); err != nil {
			return err
		}
	}
	return nil
}

func (tr *TestRunner) parallelize(args []string) error {
	ch := make(chan string, tr.Ncpu)
	ech := make(chan error, 1)

	var ewg, wg sync.WaitGroup
	var errs []error

	// start workers
	wg.Add(tr.Ncpu)
	for i := 0; i < tr.Ncpu; i++ {
		go func(wg *sync.WaitGroup) {
			for fn := range ch {
				if err := tr.runTest(fn); err != nil {
					ech <- err
				}
			}
			wg.Done()
		}(&wg)
	}

	// harvest errors
	ewg.Add(1)
	go func() {
		for e := range ech {
			errs = append(errs, e)
		}
		ewg.Done()
	}()

	// queue up work for the workers; this goroutine _will_
	for _, fn := range args {
		ch <- fn
	}
	// and tell workers that we're done
	close(ch)

	// wait for them to complete
	wg.Wait()

	// then complete harvesting all errors
	close(ech)
	ewg.Wait()

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// TestEnv captures the runtime environment of the current testsuite
type TestEnv struct {
	Lhs string
	Rhs string

	TestRoot string
	TestName string

	Log logger.Logger

	// number of concurrenct cpus to use
	Ncpu int

	Start time.Time
}

// Run a single test in file 'fn'
func (tr *TestRunner) runTest(fn string) (err error) {
	var ts []testCmd

	ts, err = readTest(fn)
	if err != nil {
		return err
	}

	// setup test env
	var env *TestEnv
	tname := filepath.Base(fn)
	env, err = tr.makeEnv(tname)
	if err != nil {
		return err
	}

	defer func(e *error) {
		if *e != nil {
			env.Log.Info("test %s complete: error:\n%s", tname, *e)
		}
		env.Log.Close()
	}(&err)

	// substitute environment vars in each arg
	lookup := map[string]string{
		"LHS":   env.Lhs,
		"RHS":   env.Rhs,
		"ROOT":  env.TestRoot,
		"TNAME": env.TestName,

		// TODO: Other vars in the future
	}

	env.Log.Info("testroot %s; starting test %s ..", env.TestRoot, env.TestName)
	for _, t := range ts {
		cmd := t.Cmd

		args := make([]string, 0, len(t.Args))
		for _, s := range t.Args[1:] {
			d := os.Expand(s, func(key string) string {
				v, ok := lookup[key]
				if !ok {
					return ""
				}
				return v
			})
			if len(d) == 0 {
				return fmt.Errorf("%s: %s: can't expand env %s", tname, cmd.Name(), s)
			}
			args = append(args, d)
		}

		if err = cmd.Run(env, args); err != nil {
			return fmt.Errorf("%s: %s: %w", tname, cmd.Name(), err)
		}
	}

	// cleanup as we go - so we don't accumulate cruft
	env.Log.Info("tests passed; removing %s ..", env.TestRoot)
	if err = os.RemoveAll(env.TestRoot); err != nil {
		return fmt.Errorf("%s: cleanup %s: %w", tname, env.TestRoot, err)
	}

	return nil
}

// make the test environment that's common to each individual test.
func (tr *TestRunner) makeEnv(tname string) (*TestEnv, error) {
	tmpdir := filepath.Join(tr.Tempdir, tname)
	lhs := filepath.Join(tmpdir, "lhs")
	rhs := filepath.Join(tmpdir, "rhs")
	logfile := filepath.Join(tmpdir, "test.log")
	if tr.LogStdout || *logStdout {
		logfile = "STDOUT"
	}

	if err := os.MkdirAll(lhs, 0700); err != nil {
		return nil, fmt.Errorf("%s: LHS: %w", tname, err)
	}

	if err := os.MkdirAll(rhs, 0700); err != nil {
		return nil, fmt.Errorf("%s: RHS: %w", tname, err)
	}

	log, err := logger.NewLogger(logfile, logger.LOG_DEBUG, tname, logger.Ldate|logger.Ltime|logger.Lmicroseconds|logger.Lfileloc)
	if err != nil {
		return nil, fmt.Errorf("%s: logfile: %w", tname, err)
	}

	e := &TestEnv{
		Lhs:      lhs,
		Rhs:      rhs,
		TestRoot: tmpdir,
		TestName: tname,
		Log:      log,
		Ncpu:     tr.Ncpu,
		Start:    time.Now(),
	}

	return e, nil
}

func (t *TestEnv) String() string {
	s := fmt.Sprintf("TestEnv: name %s: Root: %s\n\tLHS %s, RHS %s\n",
		t.TestName, t.TestRoot, t.Lhs, t.Rhs)
	return s
}


var logStdout = flag.Bool("log-stdout", false, "Send logs to stdout")
