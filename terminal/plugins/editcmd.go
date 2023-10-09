package main

import (
	"context"
	"fmt"
	"github.com/892294101/dds/dbs/sci/terminal/api"
	"os"
)

type editParamCmd string

func (e editParamCmd) Name() string {
	return string(e)
}
func (e editParamCmd) Usage() string {
	return `edit param configuration info`
}
func (e editParamCmd) ShortDesc() string {
	return `edit an process group`
}
func (e editParamCmd) LongDesc() string {
	return e.ShortDesc()
}
func (e editParamCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {
	fmt.Println("editParamCmd")
	return ctx, nil
}

// command module
type editCmds struct{}

func (t *editCmds) Init(ctx context.Context) error {
	return nil
}

func (t *editCmds) Registry() (re map[string]map[string]api.Command) {
	re = make(map[string]map[string]api.Command)
	re[LibType] = make(map[string]api.Command)
	re[LibType][api.PARAMS] = editParamCmd(api.PARAMS)

	return re
}

const LibType = "EDIT"

var Commands editCmds
