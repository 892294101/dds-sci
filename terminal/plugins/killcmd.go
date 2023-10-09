package main

import (
	"context"
	"fmt"
	"github.com/892294101/dds/dbs/sci/terminal/api"
	"os"
)

type killExtractCmd string

func (t killExtractCmd) Name() string {
	return string(t)
}
func (t killExtractCmd) Usage() string {
	return `kill EXTRACT EXT_PER LSN fileNumber:Position`
}
func (t killExtractCmd) ShortDesc() string {
	return `kill an extract group`
}
func (t killExtractCmd) LongDesc() string {
	return t.ShortDesc()
}
func (t killExtractCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {
	fmt.Println("killExtractCmd")
	return ctx, nil
}

type killReplicatCmd string

func (t killReplicatCmd) Name() string {
	return string(t)
}
func (t killReplicatCmd) Usage() string {
	return `kill replicat REP_PER LSN fileNumber:Position`
}
func (t killReplicatCmd) ShortDesc() string {
	return `kill an replicat group`
}
func (t killReplicatCmd) LongDesc() string {
	return t.ShortDesc()
}
func (t killReplicatCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {
	fmt.Println("addReplicatCmd")
	return ctx, nil
}

type killTransmitCmd string

func (t killTransmitCmd) Name() string {
	return string(t)
}
func (t killTransmitCmd) Usage() string {
	return `kill transmit TRAN_PER LSN fileNumber:Position`
}
func (t killTransmitCmd) ShortDesc() string {
	return `kill an transmit group`
}
func (t killTransmitCmd) LongDesc() string {
	return t.ShortDesc()
}
func (t killTransmitCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {
	fmt.Println("addTransmitCmd")
	return ctx, nil
}

// command module
type killCmds struct{}

func (t *killCmds) Init(ctx context.Context) error {
	return nil
}

func (t *killCmds) Registry() (re map[string]map[string]api.Command) {
	re = make(map[string]map[string]api.Command)
	re[LibType] = make(map[string]api.Command)
	re[LibType][api.EXTRACT] = killExtractCmd(api.EXTRACT)
	re[LibType][api.REPLICAT] = killReplicatCmd(api.REPLICAT)
	re[LibType][api.TRANSMIT] = killTransmitCmd(api.TRANSMIT)
	return re
}

const LibType = "KILL"

var Commands killCmds
