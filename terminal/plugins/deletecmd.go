package main

import (
	"context"
	"fmt"
	"github.com/892294101/dds/dbs/sci/terminal/api"
	"github.com/892294101/dds/dbs/sci/terminal/interactive"
	"github.com/pkg/errors"
	"os"
	"strings"
)

type deleteExtractCmd string

func (t deleteExtractCmd) Name() string {
	return string(t)
}
func (t deleteExtractCmd) Usage() string {
	return `delete EXTRACT EXT_PER`
}
func (t deleteExtractCmd) ShortDesc() string {
	return `delete an extract group`
}
func (t deleteExtractCmd) LongDesc() string {
	return t.ShortDesc()
}
func (t deleteExtractCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {
	if len(args) == 3 {
		groupId := strings.ToUpper(args[2])
		body := ctx.Value(api.RpcSubject).(*interactive.RpcBody)
		gpf, err := body.GetAllGroupAndProcessFile()
		if err != nil {
			return ctx, err
		}

		for _, info := range gpf {
			switch {
			case strings.EqualFold(groupId, info.Groups.GroupID):
				if info.Process != nil {
					return ctx, errors.Errorf("Process group is running\n")
				} else {
					if err := os.Remove(info.Groups.GroupFilePath); err != nil {
						return ctx, errors.Errorf("Process group deletion failed: %v\n", err)
					} else {
						return ctx, errors.Errorf("Process group has been deleted\n")
					}
				}
			default:
				return ctx, errors.Errorf("Process group does not exist\n")
			}
		}

	}

	return ctx, errors.Errorf("unknown parameter detected\n")
}

type deleteReplicatCmd string

func (t deleteReplicatCmd) Name() string {
	return string(t)
}
func (t deleteReplicatCmd) Usage() string {
	return `delete replicat REP_PER`
}
func (t deleteReplicatCmd) ShortDesc() string {
	return `delete an replicat group`
}
func (t deleteReplicatCmd) LongDesc() string {
	return t.ShortDesc()
}
func (t deleteReplicatCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {
	fmt.Println("addReplicatCmd")
	return ctx, nil
}

type deleteTransmitCmd string

func (t deleteTransmitCmd) Name() string {
	return string(t)
}
func (t deleteTransmitCmd) Usage() string {
	return `delete transmit TRAN_PER`
}
func (t deleteTransmitCmd) ShortDesc() string {
	return `delete an transmit group`
}
func (t deleteTransmitCmd) LongDesc() string {
	return t.ShortDesc()
}
func (t deleteTransmitCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {
	fmt.Println("addTransmitCmd")
	return ctx, nil
}

// command module
type deleteCmds struct{}

func (t *deleteCmds) Init(ctx context.Context) error {
	return nil
}

func (t *deleteCmds) Registry() (re map[string]map[string]api.Command) {
	re = make(map[string]map[string]api.Command)
	re[LibType] = make(map[string]api.Command)
	re[LibType][api.EXTRACT] = deleteExtractCmd(api.EXTRACT)
	re[LibType][api.REPLICAT] = deleteReplicatCmd(api.REPLICAT)
	re[LibType][api.TRANSMIT] = deleteTransmitCmd(api.TRANSMIT)
	return re
}

const LibType = "DELETE"

var Commands deleteCmds
