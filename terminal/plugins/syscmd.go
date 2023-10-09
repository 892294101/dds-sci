package main

import (
	"context"
	"fmt"
	"github.com/892294101/dds/dbs/sci/terminal/api"
	"github.com/892294101/dds/dbs/sci/terminal/interactive"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/mem"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// ++++++++++++++++++++++++++++++++++++++ help 命令实现
type helpCmd string

func (h helpCmd) Name() string {
	return string(h)
}

func (h helpCmd) Usage() string {
	return fmt.Sprintf("%s <command-name>", h.Name())
}

func (h helpCmd) LongDesc() string {
	return ""
}
func (h helpCmd) ShortDesc() string {
	return `prints help information for other commands.`
}

func (h helpCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {
	if ctx == nil {
		return ctx, errors.New("nil context")
	}

	out := api.GetStdout(ctx)

	cmdsVal := ctx.Value(api.ShellCommand)
	if cmdsVal == nil {
		return ctx, errors.New("nil context")
	}

	commands, ok := cmdsVal.(map[string]map[string]api.Command)

	if !ok {
		return ctx, errors.New("command map type mismatch")
	}

	initPlace := 1
	var cmdNameParam string
	if len(args) > 1 {
		cmdNameParam = strings.ToUpper(args[initPlace])
	}

	// 输出指定命令用法
	if cmdNameParam != "" {
		cmd, found := commands[cmdNameParam]
		if !found {
			return ctx, errors.New(fmt.Sprintf("Error: command not found：%v\n", *api.SliceToString(args)))
		} else {
			initPlace += 1
			if len(args) >= initPlace+1 {
				cmdName := strings.ToUpper(args[initPlace])
				cmd, found := commands[cmdNameParam][cmdName]
				if !found {
					return ctx, errors.New(fmt.Sprintf("Error: command not found：%v\n", args[initPlace]))
				} else {
					fmt.Fprintf(out, "\n%s\n", cmdNameParam)
					if cmd.Usage() != "" {
						fmt.Fprintf(out, "  Usage: %s\n", cmd.Usage())
					}
					if cmd.ShortDesc() != "" {
						fmt.Fprintf(out, "  %s\n\n", cmd.ShortDesc())
					}
				}

			} else {
				for _, cmd := range cmd {
					fmt.Fprintf(out, "\n%s\n", cmdNameParam)
					if cmd.Usage() != "" {
						fmt.Fprintf(out, "  Usage: %s\n", cmd.Usage())
					}
					if cmd.ShortDesc() != "" {
						fmt.Fprintf(out, "  %s\n\n", cmd.ShortDesc())
					}
				}
			}

		}

		return ctx, nil
	}

	fmt.Fprintf(out, "\n%s: %s\n", h.Name(), h.ShortDesc())
	fmt.Fprintln(out, "\nAvailable commands")
	fmt.Fprintln(out, "------------------")

	for types, cmdSet := range commands {
		fmt.Fprintf(out, "\nSummary of Procedure %s Commands\n", types)
		for name, command := range cmdSet {
			fmt.Fprintf(out, "%9s: %s\n", name, command.ShortDesc())
		}
	}

	fmt.Fprintf(out, "\nUse \"%s <command-name>\" for detail about the specified command\n\n", api.HELP)
	return ctx, nil
}

// ++++++++++++++++++++++++++++++++++++++ exit 命令实现
type exitCmd string

func (e exitCmd) Name() string {
	return string(e)
}

func (e exitCmd) Usage() string {
	return "exit"
}

func (e exitCmd) LongDesc() string {
	return ""
}

func (e exitCmd) ShortDesc() string {
	return `exits the interactive shell immediately`
}
func (e exitCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {
	out := api.GetStdout(ctx)
	body := ctx.Value(api.RpcSubject).(*interactive.RpcBody)
	for _, client := range body.Pro {
		if client.MD != nil {
			_ = client.MD.Close()
		}
	}
	_, _ = fmt.Fprintf(out, "\n")
	os.Exit(0)
	return ctx, nil
}

// ++++++++++++++++++++++++++++++++++++++ prompt 命令实现(它可以改变命令行的提示符)
type promptCmd string

func (p promptCmd) Name() string {
	return string(p)
}

func (p promptCmd) Usage() string {
	return fmt.Sprintf("%s <NEW-%s>", api.PROMPT, api.PROMPT)
}

func (p promptCmd) LongDesc() string {
	return ""

}

func (p promptCmd) ShortDesc() string {
	return fmt.Sprintf(`Set up a new command %s for the interactive terminal`, api.PROMPT)
}

func (p promptCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {
	if len(args) < 2 {
		return ctx, errors.Errorf("unable to set %s, see usage\n", api.PROMPT)
	}
	return context.WithValue(ctx, api.ShellPrompt, api.SetPrompt(args[1])), nil
}

// ++++++++++++++++++++++++++++++++++++++ sys 系统信息命令实现（它可以返回当前系统环境信息）
type sysinfoCmd string

func (s sysinfoCmd) Name() string {
	return string(s)
}

func (s sysinfoCmd) Usage() string {
	return s.Name()
}

func (s sysinfoCmd) LongDesc() string {
	return ""

}

func (s sysinfoCmd) ShortDesc() string {
	return `Output current system environment information`
}

func (s sysinfoCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {
	out := api.GetStdout(ctx)

	hostname, _ := os.Hostname()
	exe, _ := os.Executable()

	memStats, _ := mem.VirtualMemory()

	info := []struct{ name, value string }{
		{"Platform", runtime.GOARCH},
		{"OS", runtime.GOOS},
		{"CPU(s)", strconv.Itoa(runtime.NumCPU())},
		{"Total Memory(MB)", strconv.FormatUint(memStats.Total/1024/1024, 10)},
		{"Used Memory(MB)", strconv.FormatUint(memStats.Used/1024/1024, 10)},
		{"HostName", hostname},
		{"PageSize", strconv.Itoa(os.Getpagesize())},
		{"GroupID", strconv.Itoa(os.Getgid())},
		{"UserID", strconv.Itoa(os.Geteuid())},
		{"Pid", strconv.Itoa(os.Getpid())},
		{"BinaryFile", exe},
	}

	fmt.Fprint(out, "\nSystem Info")
	for _, k := range info {
		fmt.Fprintf(out, "\n%17s: %s", k.name, k.value)
	}
	fmt.Fprintf(out, "\n\n")
	return ctx, nil
}

// ++++++++++++++++++++++++++++++++++++++ localSys 执行本地系统命令
type localSys string

func (local localSys) Name() string { return string(local) }

func (local localSys) Usage() string { return local.Name() }

func (local localSys) LongDesc() string { return "" }

func (local localSys) ShortDesc() string {
	return `Execute local operating system commands`
}

func (local localSys) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {

	if len(args) > 1 {
		var commandSet string
		var cmdSet []string
		for i, arg := range args {
			if i == 1 {
				commandSet = arg
			} else if i > 1 {
				cmdSet = append(cmdSet, arg)
			}
		}
		/*
			cmd := exec.Command(commandSet, cmdSet...)

			stdout, err := cmd.StdoutPipe()
			if err != nil {
				fmt.Println("cmd.StdoutPipe: ", err)
				return ctx, err
			}
			cmd.Stderr = os.Stderr

			err = cmd.Start()
			if err != nil {
				return ctx, errors.Errorf("%s\n", err)
			}

			reader := bufio.NewReader(stdout)
			for {
				line, err2 := reader.ReadString('\n')
				if err2 != nil || io.EOF == err2 {
					break
				}
				fmt.Fprint(os.Stdout, line)
			}

			if err := cmd.Wait(); err != nil {
				return ctx, errors.Errorf("%s\n", strings.TrimSpace(err.Error()[strings.Index(err.Error(), ":")+1:]))
			}*/

		cmd := exec.Command(commandSet, cmdSet...)
		cmd.Stdin = ctx.Value(api.ShellStdin).(io.Reader)
		cmd.Stdout = ctx.Value(api.ShellStdout).(io.Writer)
		cmd.Stderr = ctx.Value(api.ShellStderr).(io.Writer)
		if err := cmd.Run(); err != nil {
			return ctx, errors.Errorf("%v\n", err)
		}
		fmt.Fprint(ctx.Value(api.ShellStdout).(io.Writer), "\n")
	} else {
		return ctx, errors.Errorf("Command does not meet execution conditions\n")
	}

	return ctx, nil
}

// sysCommands 表示支持命令的集合
type sysCommands struct {
	stdout io.Writer
}

func (sc *sysCommands) Init(ctx context.Context) error {
	return nil
}

func (sc *sysCommands) Registry() (re map[string]map[string]api.Command) {
	re = make(map[string]map[string]api.Command)
	re[LibType] = make(map[string]api.Command)
	re[LibType][api.HELP] = helpCmd(api.HELP)
	re[LibType][api.EXIT] = exitCmd(api.EXIT)
	re[LibType][api.PROMPT] = promptCmd(api.PROMPT)
	re[LibType][api.SYS] = sysinfoCmd(api.SYS)
	re[LibType][api.HO] = localSys(api.HO)
	return re
}

// 插件类型
const LibType = "SYS"

// 插件指针
var Commands sysCommands
