package api

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	PluginsDir    = "lib"
	CmdSymbolName = "Commands"    // 用来从插件中获取指针的type名称
	CmdSymbolType = "LibType"     // 用来从插件中获取指针的type名称
	DefaultPrompt = "CLI>"        // 默认命令提示符
	ShellPrompt   = "ShellPrompt" // 命令提示符Key
	ShellStdout   = "ShellStdout" // Stdout map的key
	ShellStderr   = "ShellStderr" // Stderr map的key
	ShellStdin    = "ShellStdin"  // Stdin map的key
	ShellCommand  = "ShellCommand"
	RpcSubject    = "RPCSUBJECT"
	LogWrite      = "LOGWRITE"
)

// 系统命令
const (
	HELP   = "HELP"
	EXIT   = "EXIT"
	PROMPT = "PROMPT"
	SYS    = "SYS"
	HO     = "HO"
)

// extract命令
const (
	PARAMS   = "PARAMS"
	EXTRACT  = "EXTRACT"
	REPLICAT = "REPLICAT"
	TRANSMIT = "TRANSMIT"
	UNIFIED  = "UNIFIED"
)

const (
	EXTSEQNO = "extseqno"
	EXTRBA   = "extrba"
	LOGNUM   = "lognum"
	LOGPOS   = "logpos"
	TRAILLOG = "traillog" // 值切勿改变其大小写
	TRANLOG  = "tranlog"  // 值切勿改变其大小写
	SCN      = "scn"      // 值切勿改变其大小写
)

// extract命令
const (
	ALL = "ALL"
)

type LibSys struct {
	Type    string
	Command string
}

// 根据执行文件路径获取程序的HOME路径
func GetHomeDir() (homeDir string) {
	file, _ := exec.LookPath(os.Args[0])
	ExecFilePath, _ := filepath.Abs(file)

	sysOs := runtime.GOOS
	switch sysOs {
	case "windows":
		execFileSlice := strings.Split(ExecFilePath, `\`)
		HomeDirectory := execFileSlice[:len(execFileSlice)-2]
		for i, v := range HomeDirectory {
			if v != "" {
				if i > 0 {
					homeDir += `\` + v
				} else {
					homeDir += v
				}
			}
		}
	case "linux", "darwin":
		execFileSlice := strings.Split(ExecFilePath, "/")
		HomeDirectory := execFileSlice[:len(execFileSlice)-2]
		for _, v := range HomeDirectory {
			if v != "" {
				homeDir += `/` + v
			}
		}
	default:
		fmt.Printf("Unknown operation type: %s\n", sysOs)
	}

	if homeDir == "" {
		fmt.Printf("Get program home directory failed: %s\n", homeDir)
	}
	return homeDir
}

func GetStdout(ctx context.Context) io.Writer {
	var out io.Writer = os.Stdout
	if ctx == nil {
		return out
	}
	if outVal := ctx.Value(ShellStdout); outVal != nil {
		if stdout, ok := outVal.(io.Writer); ok {
			out = stdout
		}
	}
	return out
}

func SetPrompt(prompt ...string) string {
	if len(prompt) == 0 {
		return fmt.Sprintf("%s", DefaultPrompt)
	}

	return fmt.Sprintf("%s>", strings.ToUpper(prompt[0]))
}

func GetPrompt(ctx context.Context) string {
	prompt := DefaultPrompt
	if ctx == nil {
		return prompt
	}
	if promptVal := ctx.Value(ShellPrompt); promptVal != nil {
		if p, ok := promptVal.(string); ok {
			prompt = p
		}
	}
	return prompt
}

// 切片转为字符类型
func SliceToString(kv []string) *string {
	var kwsb strings.Builder
	var kw string
	if len(kv) > 0 {
		for i, v := range kv {
			if i == len(kv)-1 {
				kwsb.WriteString(v)
			} else {
				kwsb.WriteString(v)
				kwsb.WriteString(" ")
			}
		}
		kw = kwsb.String()
		return &kw
	}
	return nil
}
