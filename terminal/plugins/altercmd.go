package main

import (
	"context"
	"fmt"
	"github.com/892294101/dds/dbs/metadata"
	"github.com/892294101/dds/dbs/sci/terminal/api"
	"github.com/892294101/dds/dbs/sci/terminal/interactive"
	"github.com/892294101/dds/dbs/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
)

type alterExtractCmd string

func (t alterExtractCmd) Name() string {
	return string(t)
}
func (t alterExtractCmd) Usage() string {
	return fmt.Sprintf("%v\n%9s%v\n%9s%v\n",
		"ALTER EXTRACT <process name> TRANLOG LOGNUM <logfile number> LOGPOS <logfile position>",
		"", "ALTER EXTRACT <process name> TRAILLOG EXTSEQNO <trail log number> EXTRBA <trail log rba>",
		"", "ALTER EXTRACT <process name> SCN <commit number>",
	)
}
func (t alterExtractCmd) ShortDesc() string {
	return `alter an extract group`
}
func (t alterExtractCmd) LongDesc() string {
	return t.ShortDesc()
}
func (t alterExtractCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {
	if len(args) < 3 {
		return ctx, errors.Errorf("please enter the group process name\n")
	}
	body := ctx.Value(api.RpcSubject).(*interactive.RpcBody)
	info, err := body.GetAllGroupAndProcessFile()
	if err != nil {
		return ctx, err
	}

	var exist bool
	var pi *utils.ProcessInfo
	for _, processInfo := range info {
		if strings.EqualFold(processInfo.Groups.GroupID, args[2]) {
			exist = true
			pi = &processInfo
			break
		}
	}

	if !exist {
		return ctx, errors.Errorf("process group does not exist\n")
	}
	log := ctx.Value(api.LogWrite).(*logrus.Logger)
	// 表示进程未运行
	if pi != nil && pi.Process == nil {
		apv, err := api.AttrParse(args, pi)
		if err != nil {
			return ctx, err
		}
		switch v := apv.(type) {
		case *api.LSNAttributes:
			md, err := metadata.InitMetaData(pi.Groups.GroupID, pi.Groups.DbType, pi.Groups.ProcessType, log, metadata.LOAD)
			defer md.Close()
			if err != nil {
				return ctx, err
			}
			sn, pn, err := md.GetPosition()
			if err != nil {
				return ctx, err
			}
			log.Infof("before modification: sequence %v position %v", *sn, *pn)

			if err := md.SetPosition(v.GetLSN()); err != nil {
				return ctx, err
			}
			sn, pn, err = md.GetPosition()
			if err != nil {
				return ctx, err
			}
			fmt.Fprintf(ctx.Value(api.ShellStdout).(io.Writer), "Process altered\n\n")
			log.Infof("after modification: sequence %v position %v", *sn, *pn)
		case *api.TrailAttrBody:
			md, err := metadata.InitMetaData(pi.Groups.GroupID, pi.Groups.DbType, pi.Groups.ProcessType, log, metadata.LOAD)
			defer md.Close()
			if err != nil {
				return ctx, err
			}
			sn, pn, err := md.GetFilePosition()
			if err != nil {
				return ctx, err
			}
			log.Infof("before modification: sequence %v rba %v", *sn, *pn)

			if err := md.SetFilePosition(v.GetTSN()); err != nil {
				return ctx, err
			}
			sn, pn, err = md.GetFilePosition()
			if err != nil {
				return ctx, err
			}
			fmt.Fprintf(ctx.Value(api.ShellStdout).(io.Writer), "Process altered\n\n")
			log.Infof("after modification: sequence %v rba %v", *sn, *pn)
		default:
			return ctx, errors.Errorf("unknown modified attribute\n")
		}

	} else {
		// 表示进程运行中，不可以修改
		return ctx, errors.Errorf("group process is running and cannot be modified\n")
	}

	return ctx, nil
}

type alterReplicatCmd string

func (t alterReplicatCmd) Name() string {
	return string(t)
}
func (t alterReplicatCmd) Usage() string {
	return `alter replicat REP_PER LSN fileNumber:Position`
}
func (t alterReplicatCmd) ShortDesc() string {
	return `alter an replicat group`
}
func (t alterReplicatCmd) LongDesc() string {
	return t.ShortDesc()
}
func (t alterReplicatCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {
	fmt.Println("addReplicatCmd")
	return ctx, nil
}

type alterTransmitCmd string

func (t alterTransmitCmd) Name() string {
	return string(t)
}
func (t alterTransmitCmd) Usage() string {
	return `alter transmit TRAN_PER LSN fileNumber:Position`
}
func (t alterTransmitCmd) ShortDesc() string {
	return `alter an transmit group`
}
func (t alterTransmitCmd) LongDesc() string {
	return t.ShortDesc()
}
func (t alterTransmitCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {
	fmt.Println("addTransmitCmd")
	return ctx, nil
}

// command module
type alterCmds struct{}

func (t *alterCmds) Init(ctx context.Context) error {
	return nil
}

func (t *alterCmds) Registry() (re map[string]map[string]api.Command) {
	re = make(map[string]map[string]api.Command)
	re[LibType] = make(map[string]api.Command)
	re[LibType][api.EXTRACT] = alterExtractCmd(api.EXTRACT)
	re[LibType][api.REPLICAT] = alterReplicatCmd(api.REPLICAT)
	re[LibType][api.TRANSMIT] = alterTransmitCmd(api.TRANSMIT)
	return re
}

const LibType = "ALTER"

var Commands alterCmds
