package main

import (
	"context"
	"fmt"
	"github.com/892294101/dds/dbs/sci/terminal/api"
	"os"
)

type viewParamCmd string

func (t viewParamCmd) Name() string {
	return string(t)
}
func (t viewParamCmd) Usage() string {
	return `view transmit TRAN_PER LSN fileNumber:Position`
}
func (t viewParamCmd) ShortDesc() string {
	return `view an transmit group`
}
func (t viewParamCmd) LongDesc() string {
	return t.ShortDesc()
}
func (t viewParamCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {
	fmt.Println("viewParamCmd")
	return ctx, nil
}

// command module
type viewCmds struct{}

func (t *viewCmds) Init(ctx context.Context) error {
	return nil
}

func (t *viewCmds) Registry() (re map[string]map[string]api.Command) {
	re = make(map[string]map[string]api.Command)
	re[LibType] = make(map[string]api.Command)
	re[LibType][api.PARAMS] = viewParamCmd(api.PARAMS)
	return re
}

const LibType = "VIEW"

var Commands viewCmds
