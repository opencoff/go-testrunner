// tparse.go -- lex and parse test harness

package testrunner

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/opencoff/shlex"
)

// a command is an executor of one of the test-harness commands.
type Cmd interface {
	Run(e *TestEnv, args []string) error
	Name() string
}

// singleton struct to register test harness commands.
type commands struct {
	sync.Mutex
	once sync.Once
	cmds map[string]Cmd
}

// testCmd captures the parsed contents of a test file (.t file)
type testCmd struct {
	Cmd  Cmd
	Args []string
}

func readTest(fn string) ([]testCmd, error) {
	var line string

	fd, err := os.Open(fn)
	if err != nil {
		return nil, err
	}

	defer fd.Close()

	tests := make([]testCmd, 0, 4)
	b := bufio.NewScanner(fd)
	cmds := &_Commands
	for n := 1; b.Scan(); n++ {
		part := strings.TrimSpace(b.Text())
		if len(part) == 0 || part[0] == '#' {
			continue
		}

		if part[len(part)-1] == '\\' {
			line += part[:len(part)-1]
			continue
		}

		line += part
		args, err := shlex.Split(line)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %w", fn, n, err)
		}

		line = ""
		nm := args[0]
		c, ok := cmds.cmds[nm]
		if !ok {
			return nil, fmt.Errorf("%s:%d: unknown command %s", fn, n, nm)
		}

		// remember to always give each test suite a new, clean instance
		// of the respective Cmd
		t := testCmd{
			Cmd:  c,
			Args: args,
		}
		tests = append(tests, t)
	}
	if err = b.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return tests, nil
}

// global list of all registered commands
var _Commands commands

// Register a test-harness command; called by init() functions
// in the cmd_xxx.go files.
func RegisterCommand(cmd Cmd) {
	c := &_Commands

	c.Lock()
	defer c.Unlock()

	c.once.Do(func() {
		c.cmds = make(map[string]Cmd)
	})

	nm := cmd.Name()
	if _, ok := c.cmds[nm]; ok {
		panicf("%s: command already registered", nm)
	}

	c.cmds[nm] = cmd
}
