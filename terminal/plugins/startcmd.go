package main

import (
	"context"
	"fmt"
	"github.com/892294101/dds/dbs/sci/terminal/api"
	"github.com/892294101/dds/dbs/sci/terminal/interactive"
	"github.com/892294101/dds/dbs/spfile"
	"github.com/892294101/dds/dbs/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type startExtractCmd string

func (t startExtractCmd) Name() string {
	return string(t)
}
func (t startExtractCmd) Usage() string {
	return `start EXTRACT EXT_PER LSN fileNumber:Position`
}
func (t startExtractCmd) ShortDesc() string {
	return `start an extract group`
}
func (t startExtractCmd) LongDesc() string {
	return t.ShortDesc()
}
func (t startExtractCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {
	fmt.Println("startExtractCmd")
	return ctx, nil
}

type startReplicatCmd string

func (t startReplicatCmd) Name() string {
	return string(t)
}
func (t startReplicatCmd) Usage() string {
	return `start replicat REP_PER LSN fileNumber:Position`
}
func (t startReplicatCmd) ShortDesc() string {
	return `start an replicat group`
}
func (t startReplicatCmd) LongDesc() string {
	return t.ShortDesc()
}
func (t startReplicatCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {
	fmt.Println("addReplicatCmd")
	return ctx, nil
}

type startTransmitCmd string

func (t startTransmitCmd) Name() string {
	return string(t)
}
func (t startTransmitCmd) Usage() string {
	return `start transmit TRAN_PER LSN fileNumber:Position`
}
func (t startTransmitCmd) ShortDesc() string {
	return `start an transmit group`
}
func (t startTransmitCmd) LongDesc() string {
	return t.ShortDesc()
}
func (t startTransmitCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {
	fmt.Println("addTransmitCmd")
	return ctx, nil
}

type unifiedStartCmd string

func (t unifiedStartCmd) Name() string {
	return string(t)
}
func (t unifiedStartCmd) Usage() string {
	return `start <groupName | dds* | * >`
}
func (t unifiedStartCmd) ShortDesc() string {
	return `start a process group, extract or replicat or transmit`
}
func (t unifiedStartCmd) LongDesc() string {
	return t.ShortDesc()
}
func (t unifiedStartCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {
	if len(args) > 2 {
		return ctx, errors.Errorf("unknown parameter\n")
	}

	body := ctx.Value(api.RpcSubject).(*interactive.RpcBody)
	info, err := body.GetAllGroupAndProcessFile()
	if err != nil {
		return ctx, err
	}

	dir, err := utils.GetHomeDirectory()
	if err != nil {
		return ctx, err
	}
	var execDir string
	var binaryFile string
	var proType string
	var exist bool

	// 提前验证Oracle参数文件
	log := ctx.Value(api.LogWrite).(*logrus.Logger)

	for _, processInfo := range info {
		if strings.EqualFold(processInfo.Groups.GroupID, args[1]) {
			if processInfo.Process != nil {
				return ctx, errors.Errorf("process group is running\n")
			}
			exist = true
			switch processInfo.Groups.DbType {
			case spfile.GetMySQLName():
				execDir = *dir
				binaryFile = "mysqlextract"
				proType = spfile.GetMySQLName()
			case spfile.GetOracleName():
				execDir = *dir
				binaryFile = "oracleextract"
				proType = spfile.GetOracleName()

				pfile, err := spfile.LoadSpfile(fmt.Sprintf("%s.desc", strings.ToUpper(args[1])), spfile.UTF8, log, processInfo.Groups.DbType, processInfo.Groups.ProcessType)
				if err != nil {
					return ctx, err
				}

				if err := pfile.Production(); err != nil {
					return ctx, err
				}
				// 生成的参数转为json格式，并加载到sqlite数据库，供其它进程调用
				if err := pfile.LoadToDatabase(); err != nil {
					return ctx, err
				}

				if !strings.EqualFold(*pfile.GetProcessName(), args[1]) {
					return ctx, errors.Errorf("Process name mismatch: %s", *pfile.GetProcessName())
				}

			default:
				return ctx, errors.Errorf("database type is not supported: %v\n", processInfo.Groups.DbType)
			}
			continue
		}
	}

	if !exist {
		return ctx, errors.Errorf("process group does not exist\n")
	}

	lockFile := filepath.Join(*dir, "tmp", strings.ToUpper(args[1])+".lock")
	if utils.IsFileExist(lockFile) {
		return ctx, errors.Errorf("process group is starting\n")
	}
	_, err = os.Create(lockFile)
	if err != nil {
		return ctx, errors.Errorf("lock file creation failed: %v\n", err)
	}
	defer os.Remove(lockFile)

	fp := interactive.NewForkProcess()

	switch proType {
	case spfile.GetMySQLName():
		err = fp.InitFork(execDir, binaryFile, []string{"-processid", args[1]}, log)
	case spfile.GetOracleName():
		err = fp.InitFork(execDir, binaryFile, []string{args[1]}, log)
	}

	if err != nil {
		return ctx, err
	}

	_, err = fp.Start(log)
	if err != nil {
		return ctx, fmt.Errorf("fork process failed: %v\n", err.Error())
	}

	fmt.Fprintf(ctx.Value(api.ShellStdout).(io.Writer), "Process group started\n\n")

	return ctx, nil
}

// command module
type startCmds struct{}

func (t *startCmds) Init(ctx context.Context) error {
	return nil
}

func (t *startCmds) Registry() (re map[string]map[string]api.Command) {
	re = make(map[string]map[string]api.Command)
	re[LibType] = make(map[string]api.Command)
	re[LibType][api.EXTRACT] = startExtractCmd(api.EXTRACT)
	re[LibType][api.REPLICAT] = startReplicatCmd(api.REPLICAT)
	re[LibType][api.TRANSMIT] = startTransmitCmd(api.TRANSMIT)
	re[LibType][api.UNIFIED] = unifiedStartCmd(api.UNIFIED)
	return re
}

const LibType = "START"

var Commands startCmds
