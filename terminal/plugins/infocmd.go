package main

import (
	"context"
	"fmt"
	"github.com/892294101/dds/dbs/sci/terminal/api"
	"github.com/892294101/dds/dbs/sci/terminal/interactive"
	"github.com/892294101/dds/dbs/utils"
	"github.com/pkg/errors"
	"io"
	"os"
	"strings"
)

type infoAllCmd string

func (t infoAllCmd) Name() string {
	return string(t)
}
func (t infoAllCmd) Usage() string {
	return `INFO ALL`
}
func (t infoAllCmd) ShortDesc() string {
	return `Displays status and lag for all Data Distribution Stream on a system`
}
func (t infoAllCmd) LongDesc() string {
	return t.ShortDesc()
}
func (t infoAllCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {
	body := ctx.Value(api.RpcSubject).(*interactive.RpcBody)
	info, err := body.List()
	if err != nil {
		return nil, err
	}
	_, _ = fmt.Fprintf(ctx.Value(api.ShellStdout).(io.Writer), "%v\n", *info)
	return ctx, nil
}

type infoExtractCmd string

func (t infoExtractCmd) Name() string {
	return string(t)
}
func (t infoExtractCmd) Usage() string {
	return fmt.Sprintf("%s\n%9s%s", "INFO EXTRACT DDS*", "", "INFO EXTRACT *")
}
func (t infoExtractCmd) ShortDesc() string {
	return `Returns information about an Extract group`
}
func (t infoExtractCmd) LongDesc() string {
	return t.ShortDesc()
}
func (t infoExtractCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {

	return ctx, nil
}

type infoReplicatCmd string

func (t infoReplicatCmd) Name() string {
	return string(t)
}
func (t infoReplicatCmd) Usage() string {
	return fmt.Sprintf("%s\n%9s%s", "INFO REPLICAT DDS*", "", "INFO REPLICAT *")
}
func (t infoReplicatCmd) ShortDesc() string {
	return `Returns information about an Replicat group`
}
func (t infoReplicatCmd) LongDesc() string {
	return t.ShortDesc()
}
func (t infoReplicatCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {
	fmt.Println("infoReplicatCmd")
	return ctx, nil
}

type infoTransmitCmd string

func (t infoTransmitCmd) Name() string {
	return string(t)
}
func (t infoTransmitCmd) Usage() string {
	return fmt.Sprintf("%s\n%9s%s", "INFO TRANSMIT DDS*", "", "INFO TRANSMIT *")
}
func (t infoTransmitCmd) ShortDesc() string {
	return `Returns information about an Transmit group`
}
func (t infoTransmitCmd) LongDesc() string {
	return t.ShortDesc()
}
func (t infoTransmitCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {

	return ctx, nil
}

type unifiedInfoCmd string

func (t unifiedInfoCmd) Name() string {
	return string(t)
}
func (t unifiedInfoCmd) Usage() string {
	return fmt.Sprintf("%s\n%9s%s", "INFO TRANSMIT DDS*", "", "INFO TRANSMIT *")
}
func (t unifiedInfoCmd) ShortDesc() string {
	return `Returns information about an Transmit group`
}
func (t unifiedInfoCmd) LongDesc() string {
	return t.ShortDesc()
}
func (t unifiedInfoCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {
	if len(args) > 2 {
		return ctx, errors.Errorf("unknown parameter\n")
	}

	body := ctx.Value(api.RpcSubject).(*interactive.RpcBody)
	info, err := body.GetAllGroupAndProcessFile()
	if err != nil {
		return ctx, err
	}

	var exist bool
	var pi *utils.ProcessInfo
	for _, processInfo := range info {
		if strings.EqualFold(processInfo.Groups.GroupID, args[1]) {
			exist = true
			pi = &processInfo
			break
		}
	}

	if !exist {
		return ctx, errors.Errorf("process group does not exist\n")
	}

	detail, err := body.Detail(pi)
	if err != nil {
		return nil, err
	}
	_, _ = fmt.Fprintf(ctx.Value(api.ShellStdout).(io.Writer), "%v\n", *detail)
	return ctx, nil
}

// command module
type infoAllCmds struct{}

func (t *infoAllCmds) Init(ctx context.Context) error {
	return nil
}

func (t *infoAllCmds) Registry() (re map[string]map[string]api.Command) {
	re = make(map[string]map[string]api.Command)
	re[LibType] = make(map[string]api.Command)
	re[LibType][api.ALL] = infoAllCmd(api.ALL)
	re[LibType][api.EXTRACT] = infoExtractCmd(api.EXTRACT)
	re[LibType][api.REPLICAT] = infoReplicatCmd(api.REPLICAT)
	re[LibType][api.TRANSMIT] = infoTransmitCmd(api.TRANSMIT)
	re[LibType][api.UNIFIED] = unifiedInfoCmd(api.UNIFIED)
	return re
}

const LibType = "INFO"

var Commands infoAllCmds
