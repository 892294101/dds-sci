package main

import (
	"context"
	"fmt"
	"github.com/892294101/dds/dbs/metadata"
	"github.com/892294101/dds/dbs/sci/terminal/api"
	"github.com/892294101/dds/dbs/spfile"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
	"time"
)

type addExtractCmd string

func (t addExtractCmd) Name() string {
	return string(t)
}
func (t addExtractCmd) Usage() string {
	return `ADD EXTRACT EXT_PER LSN fileNumber:Position`
}
func (t addExtractCmd) ShortDesc() string {
	return `Creates an extract group"`
}
func (t addExtractCmd) LongDesc() string {
	return t.ShortDesc()
}
func (t addExtractCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {
	if len(args) == 4 {
		groupId := strings.ToUpper(args[2])
		if len(groupId) > 12 {
			return ctx, errors.Errorf("name cannot exceed 12 characters: %v\n", groupId)
		}

		gf := api.NewGroupFile()
		var dbType string
		switch {
		case strings.EqualFold(spfile.GetMySQLName(), args[3]):
			dbType = spfile.GetMySQLName()
			break
		case strings.EqualFold(spfile.GetOracleName(), args[3]):
			dbType = spfile.GetOracleName()
			break
		default:
			return ctx, errors.Errorf("Database type is not supported: %v\n", args[3])
		}

		err := gf.InsertGroupInfo(groupId, dbType, spfile.GetExtractName())
		if err != nil {
			return ctx, errors.Errorf("init process group data error: %v\n", err)
		}
		// 当写入组文件信息时，如果组已经存在，则返回错误
		err = gf.WriteTo()
		if err != nil {
			return ctx, err
		}

		log := ctx.Value(api.LogWrite).(*logrus.Logger)
		md, err := metadata.InitMetaData(groupId, dbType, spfile.GetExtractName(), log, metadata.CREATE)
		if err != nil {
			gf.Remove()
			return ctx, err
		}
		// 创建完成后，当函数退出时，需要把检查点文件句柄关闭掉，并解除文件锁。
		defer md.Close()
		// 初始化元数据中的创建时间
		if err := md.SetCreateTime(uint64(time.Now().Unix())); err != nil {
			gf.Remove()
			return ctx, err
		}

		if err := md.SetTransactionBeginTime(uint64(time.Now().Unix())); err != nil {
			gf.Remove()
			return ctx, err
		}

		// 初始化元数据中的数据库类型
		if err := md.SetDataBaseType(dbType); err != nil {
			gf.Remove()
			return ctx, err
		}

		// 初始化元数据中的进程类型
		if err := md.SetProcessType(spfile.GetExtractName()); err != nil {
			gf.Remove()
			return ctx, err
		}

		fmt.Fprintf(ctx.Value(api.ShellStdout).(io.Writer), "%s %s process has been added\n\n", groupId, spfile.GetExtractName())
		return ctx, nil
	}

	return ctx, errors.Errorf("missing required parameter\n")
}

type addReplicatCmd string

func (t addReplicatCmd) Name() string {
	return string(t)
}
func (t addReplicatCmd) Usage() string {
	return `ADD replicat REP_PER LSN fileNumber:Position`
}
func (t addReplicatCmd) ShortDesc() string {
	return `Creates an replicat group"`
}
func (t addReplicatCmd) LongDesc() string {
	return t.ShortDesc()
}
func (t addReplicatCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {

	return ctx, nil
}

type addTransmitCmd string

func (t addTransmitCmd) Name() string {
	return string(t)
}
func (t addTransmitCmd) Usage() string {
	return `ADD TRANSMIT TRAN_PER LSN fileNumber:Position`
}
func (t addTransmitCmd) ShortDesc() string {
	return `Creates an transmit group"`
}
func (t addTransmitCmd) LongDesc() string {
	return t.ShortDesc()
}
func (t addTransmitCmd) Exec(ctx context.Context, args []string, sgl chan os.Signal) (context.Context, error) {

	return ctx, nil
}

// command module
type addCmds struct{}

func (t *addCmds) Init(ctx context.Context) error {
	return nil
}

func (t *addCmds) Registry() (re map[string]map[string]api.Command) {
	re = make(map[string]map[string]api.Command)
	re[LibType] = make(map[string]api.Command)
	re[LibType][api.EXTRACT] = addExtractCmd(api.EXTRACT)
	re[LibType][api.REPLICAT] = addReplicatCmd(api.REPLICAT)
	re[LibType][api.TRANSMIT] = addTransmitCmd(api.TRANSMIT)
	return re
}

const LibType = "ADD"

var Commands addCmds
