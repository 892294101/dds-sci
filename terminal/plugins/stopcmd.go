package main

import (
	"context"
	"fmt"
	"github.com/892294101/dds/dbs/sci/terminal/api"
	"github.com/892294101/dds/dbs/sci/terminal/interactive"
	"github.com/892294101/dds/dbs/spfile"
	"github.com/pkg/errors"
	"io"
	"os"
	"strings"
)

type stopExtractCmd string

func (t stopExtractCmd) Name() string {
	return string(t)
}
func (t stopExtractCmd) Usage() string {
	return `stop EXTRACT EXT_PER LSN fileNumber:Position`
}
func (t stopExtractCmd) ShortDesc() string {
	return `stop an extract group`
}
func (t stopExtractCmd) LongDesc() string {
	return t.ShortDesc()
}
func (t stopExtractCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {
	fmt.Println("stopExtractCmd")
	return ctx, nil
}

type stopReplicatCmd string

func (t stopReplicatCmd) Name() string {
	return string(t)
}
func (t stopReplicatCmd) Usage() string {
	return `stop replicat REP_PER LSN fileNumber:Position`
}
func (t stopReplicatCmd) ShortDesc() string {
	return `stop an replicat group`
}
func (t stopReplicatCmd) LongDesc() string {
	return t.ShortDesc()
}
func (t stopReplicatCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {
	fmt.Println("addReplicatCmd")
	return ctx, nil
}

type stopTransmitCmd string

func (t stopTransmitCmd) Name() string {
	return string(t)
}
func (t stopTransmitCmd) Usage() string {
	return `stop transmit TRAN_PER LSN fileNumber:Position`
}
func (t stopTransmitCmd) ShortDesc() string {
	return `stop an transmit group`
}
func (t stopTransmitCmd) LongDesc() string {
	return t.ShortDesc()
}
func (t stopTransmitCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {
	fmt.Println("addTransmitCmd")
	return ctx, nil
}

type unifiedStopCmd string

func (t unifiedStopCmd) Name() string {
	return string(t)
}
func (t unifiedStopCmd) Usage() string {
	return `stop <groupName | dds* | * >`
}
func (t unifiedStopCmd) ShortDesc() string {
	return `stop a process group, extract or replicat or transmit`
}
func (t unifiedStopCmd) LongDesc() string {
	return t.ShortDesc()
}
func (t unifiedStopCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {

	body := ctx.Value(api.RpcSubject).(*interactive.RpcBody)
	info, err := body.GetAllGroupAndProcessFile()
	if err != nil {
		return ctx, err
	}

	var exist bool
	for _, processInfo := range info {
		if strings.EqualFold(processInfo.Groups.GroupID, args[1]) {
			exist = true
			switch processInfo.Groups.DbType {
			case spfile.GetMySQLName():
				if processInfo.Process == nil {
					return ctx, errors.Errorf("process group is not running\n")
				}

				rpc, err := body.ConnectToServer(processInfo)
				if err != nil {
					return ctx, err
				}

				_, err = rpc.Stop()
				_ = rpc.StopRpc()
				if err != nil {
					return ctx, err
				}

				fmt.Fprintf(ctx.Value(api.ShellStdout).(io.Writer), "process group is stopping\n\n")

			default:
				return ctx, errors.Errorf("database type is not supported: %v\n", processInfo.Groups.DbType)
			}
			continue
		}
	}

	if !exist {
		return ctx, errors.Errorf("process group does not exist\n")
	}
	return ctx, nil
}

// command module
type stopCmds struct{}

func (t *stopCmds) Init(ctx context.Context) error {
	return nil
}

func (t *stopCmds) Registry() (re map[string]map[string]api.Command) {
	re = make(map[string]map[string]api.Command)
	re[LibType] = make(map[string]api.Command)
	re[LibType][api.EXTRACT] = stopExtractCmd(api.EXTRACT)
	re[LibType][api.REPLICAT] = stopReplicatCmd(api.REPLICAT)
	re[LibType][api.TRANSMIT] = stopTransmitCmd(api.TRANSMIT)
	re[LibType][api.UNIFIED] = unifiedStopCmd(api.UNIFIED)
	return re
}

const LibType = "STOP"

var Commands stopCmds
