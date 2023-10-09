package terminal

import (
	"bufio"
	"context"
	"fmt"
	"github.com/892294101/dds/dbs/ddslog"
	"github.com/892294101/dds/dbs/sci/terminal/api"
	"github.com/892294101/dds/dbs/sci/terminal/interactive"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"plugin"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
)

var Organization = "Data Distribution Stream Command Interpreter"
var Version = "UNKNOWN"
var BDate = "UNKNOWN"

var (
	reCmd = regexp.MustCompile(`\S+`)
)

type shell struct {
	ctx        context.Context
	pluginsDir string
	commands   map[string]map[string]api.Command
	closed     chan struct{}
}

// newShell returns a new shell
func newShell() *shell {
	return &shell{
		pluginsDir: api.PluginsDir,
		commands:   make(map[string]map[string]api.Command),
		closed:     make(chan struct{}),
	}
}

// Init initializes the shell with the given context
func (gosh *shell) Init(ctx context.Context) error {
	gosh.ctx = ctx
	return gosh.loadCommands()
}

func (gosh *shell) loadCommands() error {

	if _, err := os.Stat(filepath.Join(api.GetHomeDir(), gosh.pluginsDir)); err != nil {
		return err
	}

	plugins, err := listFiles(filepath.Join(api.GetHomeDir(), gosh.pluginsDir), `.*_command.so`)
	if err != nil {
		return err
	}

	for _, cmdPlugin := range plugins {
		plug, err := plugin.Open(path.Join(filepath.Join(api.GetHomeDir(), gosh.pluginsDir), cmdPlugin.Name()))
		if err != nil {
			fmt.Printf("failed to open plugin %s: %v\n", cmdPlugin.Name(), err)
			continue
		}
		cmdSymbol, err := plug.Lookup(api.CmdSymbolName)
		if err != nil {
			fmt.Printf("plugin %s does not export symbol \"%s\"\n",
				cmdPlugin.Name(), api.CmdSymbolName)
			continue
		}
		commands, ok := cmdSymbol.(api.Commands)
		if !ok {
			fmt.Printf("Symbol %s (from %s) does not implement Commands interface\n", api.CmdSymbolName, cmdPlugin.Name())
			continue
		}
		if err := commands.Init(gosh.ctx); err != nil {
			fmt.Printf("%s initialization failed: %v\n", cmdPlugin.Name(), err)
			continue
		}
		for name, cmd := range commands.Registry() {
			gosh.commands[name] = cmd
		}
		gosh.ctx = context.WithValue(gosh.ctx, api.ShellCommand, gosh.commands)
	}
	return nil
}

// other interaction based on shell environment
func (gosh *shell) otherInteraction(in *os.File, sgl chan os.Signal) {
	loopCtx := gosh.ctx
	line := make(chan string)
	pi := bufio.NewReader(in)

	ctx := context.WithValue(context.Background(), "done", "bye")
	// start a goroutine to get input from the text
	go func(input chan<- string) {
		for {
			line, err := pi.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					c, canal := context.WithCancel(ctx)
					canal()
					ctx = c
					return
				}
			}
			input <- strings.TrimSpace(line)
		}
	}(line)

	for {
		select {
		case <-sgl:
			gosh.CloseRpc()
			close(gosh.closed)
			return
		case <-ctx.Done():
			//fmt.Fprintf(loopCtx.Value(api.ShellStdout).(io.Writer), "%v\n", ctx.Value("done"))
			//_ = <-ctx.Done()
			_ = ctx.Value("done")
			gosh.CloseRpc()
			close(gosh.closed)
			return
		case input := <-line:
			var err error
			loopCtx, err = gosh.handle(loopCtx, input, sgl)
			if err != nil {
				_, _ = fmt.Fprintf(loopCtx.Value(api.ShellStderr).(io.Writer), "%v\n\n", err)
			}

		}
	}

}

func (gosh *shell) CloseRpc() {
	body := gosh.ctx.Value(api.RpcSubject).(*interactive.RpcBody)
	fmt.Println("CloseRpc")

	body.Close()
}

// Pipe interaction based on shell environment
func (gosh *shell) pipeInteraction(in *os.File, sgl chan os.Signal) {
	loopCtx := gosh.ctx
	line := make(chan string)
	pi := bufio.NewScanner(in)
	// start a goroutine to get input from the pipe
	go func(ctx context.Context, input chan<- string) {
		for pi.Scan() {
			input <- pi.Text()
			return
		}
	}(loopCtx, line)

	select {
	case <-gosh.ctx.Done():
		gosh.CloseRpc()
		close(gosh.closed)
		return
	case input := <-line:
		var err error
		loopCtx, err = gosh.handle(loopCtx, input, sgl)
		if err != nil {
			_, _ = fmt.Fprintf(loopCtx.Value(api.ShellStderr).(io.Writer), "%v\n", err)
		}
	}
}

// Open opens the shell for the given os.file
func (gosh *shell) Open(in *os.File, pipeState os.FileInfo, sgl chan os.Signal) {

	if (pipeState.Mode() & os.ModeNamedPipe) == os.ModeNamedPipe {
		gosh.pipeInteraction(in, sgl)
	} else if (pipeState.Mode() & os.ModeDevice) == os.ModeDevice {
		gosh.UserInteraction(in, sgl)
	} else {
		gosh.otherInteraction(in, sgl)
		/*loopCtx := gosh.ctx
		_, _ = fmt.Fprintf(loopCtx.Value(api.ShellStdout).(io.Writer), "This mode is not supported temporarily\n\n")*/
	}
}

// User interaction based on shell environment
func (gosh *shell) UserInteraction(in *os.File, sgl chan os.Signal) {
	loopCtx := gosh.ctx
	line := make(chan string)
	r := bufio.NewReader(in)
	for {
		// start a goroutine to get input from the user
		go func(ctx context.Context, input chan<- string) {
			for {
				// TODO: future enhancement is to capture input key by key
				// to give command granular notification of key events.
				// This could be used to implement command autocompletion.
				_, _ = fmt.Fprintf(ctx.Value(api.ShellStdout).(io.Writer), "%s ", api.GetPrompt(loopCtx))
				line, err := r.ReadString('\n')
				if err != nil {
					_, _ = fmt.Fprintf(ctx.Value(api.ShellStderr).(io.Writer), "%v\n", err)
					continue
				}
				input <- line
				return
			}
		}(loopCtx, line)

		// wait for input or cancel
		select {
		case <-gosh.ctx.Done():
			gosh.CloseRpc()
			close(gosh.closed)
			return
		case input := <-line:
			// 捕获终止信号，并丢弃
			if len(sgl) > 0 {
				var have int
				for {
					sig := <-sgl
					switch sig {
					case syscall.SIGINT:
						have += 1
					case syscall.SIGTSTP:
						have += 1
					case syscall.SIGQUIT:
						have += 1
					}
					if len(sgl) == 0 {
						break
					}
				}
				if have > 0 {
					_, _ = fmt.Fprintf(loopCtx.Value(api.ShellStdout).(io.Writer), "\n")
				}

			} else {
				var err error
				loopCtx, err = gosh.handle(loopCtx, input, sgl)
				if err != nil {
					_, _ = fmt.Fprintf(loopCtx.Value(api.ShellStderr).(io.Writer), "%v\n", err)
				}
			}
		}
	}
}

// Closed returns a channel that closes when the shell has closed
func (gosh *shell) Closed() <-chan struct{} {
	return gosh.closed
}

func (gosh *shell) handle(ctx context.Context, cmdLine string, sgl chan os.Signal) (context.Context, error) {
	line := strings.TrimSpace(cmdLine)
	if line == "" {
		_, _ = fmt.Fprint(ctx.Value(api.ShellStdout).(io.Writer), "\n")
		return ctx, nil
	}
	args := reCmd.FindAllString(line, -1)
	if args != nil {
		FirstCmdName := strings.ToUpper(args[0])
		var NotFound bool
		var timeout = time.Second * 15
		var err error
		for masterCmd := range gosh.commands {
			// help命令由于在sys包中，所以需要单独处理
			cmd, ok := gosh.commands[masterCmd][FirstCmdName]
			if ok {
				NotFound = true
				done := make(chan int, 1)
				go func(c context.Context, e *error, d chan int) {
					c, *e = cmd.Exec(c, args, sgl)
					done <- 1
				}(ctx, &err, done)

				select {
				case <-done:
					return ctx, err
				case <-sgl:
					_, _ = fmt.Fprintf(ctx.Value(api.ShellStdout).(io.Writer), "\n")
					return ctx, err
				}
			}

		}
		if len(args) > 1 {
			TwoCmdName := strings.ToUpper(args[1])
			var cmd api.Command
			var ok bool
			cmd, ok = gosh.commands[FirstCmdName][TwoCmdName]
			if !ok {
				cmd, ok = gosh.commands[FirstCmdName][api.UNIFIED]
			}

			if ok {
				NotFound = true
				done := make(chan int, 1)
				go func(c context.Context, e *error, d chan int) {
					c, *e = cmd.Exec(c, args, sgl)
					done <- 1
				}(ctx, &err, done)

				select {
				case <-done:
					return ctx, err
				case <-time.After(timeout):
					_, _ = fmt.Fprintf(ctx.Value(api.ShellStderr).(io.Writer), "Error: timed out\n\n")
					return ctx, err
				case <-sgl:
					_, _ = fmt.Fprintf(ctx.Value(api.ShellStdout).(io.Writer), "\n\n")
					return ctx, err
				}
			}
		}

		if !NotFound {
			return ctx, errors.New(fmt.Sprintf("Error: Invalid command: %s\n", *api.SliceToString(args)))
		}
	}
	return ctx, errors.New(fmt.Sprintf("Error: unable to parse command line: %s", line))
}

func listFiles(dir, pattern string) ([]os.FileInfo, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	filteredFiles := []os.FileInfo{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		matched, err := regexp.MatchString(pattern, file.Name())
		if err != nil {
			return nil, err
		}
		if matched {
			filteredFiles = append(filteredFiles, file)
		}
	}
	return filteredFiles, nil
}

func OpenShell() error {
	p := true
	if len(os.Args) > 1 {
		b := os.Args[1]
		switch {
		case strings.EqualFold(b, "-n"):
			p = false
		default:
			return errors.Errorf("This value is not supported: %v\n\n", os.Args[1])
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log, err := ddslog.InitDDSlog("CLI")
	if err != nil {
		return errors.Errorf("%s", err)
	}

	ctx = context.WithValue(ctx, api.ShellPrompt, api.SetPrompt())               // 终端符号
	ctx = context.WithValue(ctx, api.ShellStdout, os.Stdout)                     // 终端stdout
	ctx = context.WithValue(ctx, api.ShellStderr, os.Stderr)                     // 终端stderr
	ctx = context.WithValue(ctx, api.ShellStdin, os.Stdin)                       // 终端stdin
	ctx = context.WithValue(ctx, api.RpcSubject, interactive.InitRpcClient(log)) // 终端传入的RpcSubject
	ctx = context.WithValue(ctx, api.LogWrite, log)                              // 终端传入的RpcSubject

	shell := newShell()
	if err := shell.Init(ctx); err != nil {
		return errors.Errorf("\n\nfailed to initialize:\n", err)
	}

	cmdCount := len(shell.commands)
	var libraryNum int
	var cmdNum int
	if cmdCount > 0 {
		for _, m := range shell.commands {
			libraryNum += 1
			cmdNum = cmdNum + len(m)
		}

		//fmt.Printf("\nLoad %d command library, %d commands in total...\n", libraryNum, cmdNum)
		if p {
			_, _ = fmt.Fprintf(os.Stdout, "\n%s\n", Organization)
			_, _ = fmt.Fprintf(os.Stdout, "Version %s on %s\n\n", Version, BDate)
			_, _ = fmt.Fprintf(os.Stdout, "Type help for available commands\n")
		}
		_, _ = fmt.Fprintf(os.Stdout, "\n")
	} else {
		return errors.Errorf("Command set not found\n\n")
	}

	// 捕获终止信号
	SignalSet := make(chan os.Signal, 3)
	signal.Notify(SignalSet, syscall.SIGINT, syscall.SIGTSTP, syscall.SIGQUIT)

	in := os.Stdin
	pipe, err := in.Stat()
	if err != nil {
		return errors.Errorf("stdin error: %s ", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func(w *sync.WaitGroup, sgl chan os.Signal) {
		defer w.Done()
		shell.Open(in, pipe, sgl)
	}(&wg, SignalSet)

	wg.Wait()
	return nil
}
