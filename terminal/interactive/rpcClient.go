package interactive

import (
	"fmt"
	"github.com/892294101/dds/dbs/metadata"
	"github.com/892294101/dds/utils"

	"github.com/892294101/dds-sci/terminal/api"
	"github.com/892294101/ddsrpc"
	"github.com/pingcap/errors"
	"github.com/sirupsen/logrus"

	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type InfoAll struct {
	Type,
	Program,
	Status,
	GroupID,
	Lag,
	SinceChkpt string
}

type InfoAllStatistics struct {
	InfoDisplay []*InfoAll
}

func NewInfoAllDisplay() *InfoAllStatistics {
	is := new(InfoAllStatistics)
	iad := new(InfoAll)
	iad.Type = "Type"
	iad.Program = "Program"
	iad.Status = "Status"
	iad.GroupID = "Group"
	iad.Lag = "Lag"
	iad.SinceChkpt = "Time Since Chkpt"
	is.InfoDisplay = append(is.InfoDisplay, iad)
	return is
}

func (i *InfoAllStatistics) ToString() string {
	var buf strings.Builder
	for ind, info := range i.InfoDisplay {
		if ind == 0 {
			buf.WriteString(fmt.Sprintf("\n%-10s%-10s%-9s%-12s%-14s%-16s\n", info.Type, info.Program, info.Status, info.GroupID, info.Lag, info.SinceChkpt))
		} else {
			buf.WriteString(fmt.Sprintf("%-10s%-10s%-9s%-12s%-14s%-16s\n", info.Type, info.Program, info.Status, info.GroupID, info.Lag, info.SinceChkpt))
		}
	}
	return buf.String()
}

func (i *InfoAllStatistics) Set(info *InfoAll) {
	i.InfoDisplay = append(i.InfoDisplay, info)
}

type InfoDetail struct {
	Program, //  进程类型
	GroupID, //  进程ID
	Dbtype, // 数据库类型
	LastStarted, // 进程最后启动时间
	Lag, // 进程lag时间
	SinceChkpt, // 检查点延迟时间
	Status, // 进程运行状态
	FileNumber, // trail文件序列号
	FileOffset, // trail文件rba
	LogNumber, // 日志文件号
	LogOffset string // 日志文件rba
}

type InfoDetailStatistics struct {
	InfoDetailDisplay []*InfoDetail
}

func NewInfoDetailDisplay() *InfoDetailStatistics {
	is := new(InfoDetailStatistics)
	iad := new(InfoDetail)
	iad.Program = "Program"
	iad.GroupID = "GroupID"
	iad.Dbtype = "Type"
	iad.LastStarted = "Last Started"
	iad.Lag = "Lag"
	iad.SinceChkpt = "Time Since Chkpt"
	iad.Status = "Status"
	iad.FileNumber = "File Number"
	iad.FileOffset = "File Offset"
	iad.LogNumber = "Log Number"
	iad.LogOffset = "Log Offset"

	is.InfoDetailDisplay = append(is.InfoDetailDisplay, iad)
	return is
}

func (i *InfoDetailStatistics) ToString() string {
	var buf strings.Builder
	if len(i.InfoDetailDisplay) == 2 {
		buf.WriteString(fmt.Sprintf("\n%-13s%-12s%-13s%-34s%-8s%-7s\n", i.InfoDetailDisplay[1].Program, i.InfoDetailDisplay[1].GroupID, i.InfoDetailDisplay[0].LastStarted, i.InfoDetailDisplay[1].LastStarted, i.InfoDetailDisplay[0].Status, i.InfoDetailDisplay[1].Status))
		buf.WriteString(fmt.Sprintf("%-13s%-12s%-12s%12s\n", i.InfoDetailDisplay[0].Lag, i.InfoDetailDisplay[1].Lag, i.InfoDetailDisplay[0].SinceChkpt, i.InfoDetailDisplay[1].SinceChkpt))
		buf.WriteString(fmt.Sprintf("%-13s%-12s%-13s%-12s\n", i.InfoDetailDisplay[0].FileNumber, i.InfoDetailDisplay[1].FileNumber, i.InfoDetailDisplay[0].FileOffset, i.InfoDetailDisplay[1].FileOffset))
		buf.WriteString(fmt.Sprintf("%-13s%-12s%-13s%-12s\n", i.InfoDetailDisplay[0].LogNumber, i.InfoDetailDisplay[1].LogNumber, i.InfoDetailDisplay[0].LogOffset, i.InfoDetailDisplay[1].LogOffset))
	}
	return buf.String()
}

func (i *InfoDetailStatistics) Set(detail *InfoDetail) {
	i.InfoDetailDisplay = append(i.InfoDetailDisplay, detail)
}

type Processor struct {
	// RPC *ddsrpc.RpcClient // Communication body
	MD metadata.MetaData // 进程元数据
	metadata.MdHandle
}

type RpcBody struct {
	Pro map[string]*Processor
	log *logrus.Logger
}

func (c *RpcBody) ConnectToServer(pi utils.ProcessInfo) (*ddsrpc.RpcClient, error) {
	if pi.Process == nil {
		return nil, errors.Errorf("failed to acquire rpc port")
	}

	pid, err := strconv.Atoi(pi.Process.PID)
	if err != nil {
		return nil, err
	}
	if ok := utils.CheckPid(pid); ok {
		RPC, err := ddsrpc.NewRpcClient(pi.Process.PORT)
		if err != nil {
			return nil, errors.Errorf("failed to establish rpc")
		}
		return RPC, nil
	}
	return nil, errors.Errorf("process group does not exist and may have exited")
}

/*
todo 停止状态赋值要检测pid是否存在

*/

func (c *RpcBody) GetAllGroupAndProcessFile() (pfi []utils.ProcessInfo, err error) {
	home, err := utils.GetHomeDirectory()
	if err != nil {
		return nil, err
	}

	// 获取进程组信息文件信息
	var ginfo []*utils.GroupInfo
	gHome := path.Join(*home, "group")
	groupIds, err := utils.GetAllGroupFileName(gHome, "")
	if err != nil {
		return nil, err
	}
	for _, groupId := range groupIds {
		var gi utils.GroupInfo
		gf := api.NewGroupFile()
		pgi, err := gf.ReadGroupFileInfo(groupId)
		if err != nil {
			return nil, err
		}
		gi.DbType = pgi.DBType
		gi.ProcessType = pgi.ProcessType
		gi.GroupID = groupId

		gi.GroupFilePath = path.Join(*home, "group", groupId)
		ginfo = append(ginfo, &gi)
	}

	// 获取检查点文件路径和提取进程文件信息
	for _, info := range ginfo {
		ProFile := filepath.Join(*home, "pcs", info.GroupID)
		ok := utils.IsFileExist(ProFile) // 查看进程文件是否存在
		if ok {
			v, err := utils.ReadLine(ProFile)
			if err != nil {
				return nil, err
			}
			pab, err := utils.GetProcessAttribute(v)
			if err != nil {
				return nil, err
			}
			pab.File = ProFile
			pfi = append(pfi, utils.ProcessInfo{Groups: info, CheckPointFilePath: filepath.Join(*home, "chk", fmt.Sprintf("%s.%s", info.GroupID, "ce")), Process: pab})
		} else {
			pfi = append(pfi, utils.ProcessInfo{Groups: info, CheckPointFilePath: filepath.Join(*home, "chk", fmt.Sprintf("%s.%s", info.GroupID, "ce"))})
		}
	}
	return pfi, nil
}

func (c *RpcBody) Close() {
	for _, client := range c.Pro {
		if client.MD != nil {
			if err1 := client.MD.Close(); err1 != nil {
				c.log.Warnf("%v", err1)
			}
		}
	}
}

func (c *RpcBody) Detail(pi *utils.ProcessInfo) (*string, error) {
	ib := NewInfoDetailDisplay()
	md, err := metadata.InitMetaData(pi.Groups.GroupID, pi.Groups.DbType, pi.Groups.ProcessType, c.log, metadata.LOAD)
	if err != nil {
		return nil, err
	}

	var Lag string
	// 获取检查点中事务的最后一次更新时间
	bt, err := md.GetLastUpdateTime()
	if err != nil {
		return nil, err
	}

	var proState string
	if pi.Process == nil {
		proState = utils.STOPPED
	} else {
		proState = utils.RUNNING
	}

	var TimeSinceChkpt string
	checkPointLag, _ := md.GetTransactionBeginTime()
	if checkPointLag == 0 {
		Lag = "00:00:00:00"
		TimeSinceChkpt = "00:00:00:00"
	} else {
		if *bt == 0 {
			ct, err := md.GetCreateTime()
			if err != nil {
				return nil, err
			}
			Lag = utils.DataStreamLagTime(*ct)
			TimeSinceChkpt = Lag
		} else {
			Lag = utils.DataStreamLagTime(*bt)
			TimeSinceChkpt = utils.DataStreamLagTime(checkPointLag)
		}

	}
	fn, offset, err := md.GetFilePosition()
	if err != nil {
		return nil, err
	}
	pn, rba, err := md.GetPosition()
	if err != nil {
		return nil, err
	}

	var LastStarted string
	st, _ := md.GetStartTime()
	if st > 0 {
		LastStarted = utils.NanoSecondConvertToTime(st)
	}

	id := new(InfoDetail)
	id.Program = pi.Groups.ProcessType
	id.GroupID = pi.Groups.GroupID
	id.Dbtype = pi.Groups.DbType
	id.LastStarted = LastStarted
	id.Lag = Lag
	id.SinceChkpt = TimeSinceChkpt
	id.Status = proState
	id.FileNumber = strconv.Itoa(int(*fn))
	id.FileOffset = strconv.Itoa(int(*offset))
	id.LogNumber = strconv.Itoa(int(*pn))
	id.LogOffset = strconv.Itoa(int(*rba))
	ib.Set(id)
	md.Close()

	res := ib.ToString()
	return &res, nil
}
func (c *RpcBody) List() (*string, error) {
	// 获取所有进程检查点文件和进程文件
	v, err := c.GetAllGroupAndProcessFile()
	if err != nil {
		return nil, err
	}
	ib := NewInfoAllDisplay()

	for _, info := range v {
		// 打开检查点文件
		_, ok := c.Pro[info.Groups.GroupID]
		if !ok {
			// 如果检查点文件不存在，则打开
			md, err := metadata.InitMetaData(info.Groups.GroupID, info.Groups.DbType, info.Groups.ProcessType, c.log, metadata.LOAD)
			if err != nil {
				return nil, err
			}
			c.Pro[info.Groups.GroupID] = new(Processor)
			c.Pro[info.Groups.GroupID].MD = md
		}
	}

	for _, info := range v {
		_, ok := c.Pro[info.Groups.GroupID]
		if ok {
			md := c.Pro[info.Groups.GroupID].MD
			var Lag string
			// 获取检查点中事务的最后一次更新时间
			bt, err := md.GetLastUpdateTime()
			if err != nil {
				return nil, err
			}

			var proState string
			if info.Process == nil {
				proState = utils.STOPPED
			} else {
				proState = utils.RUNNING
			}

			var TimeSinceChkpt string
			checkPointLag, _ := md.GetTransactionBeginTime()
			if checkPointLag == 0 {
				Lag = "00:00:00:00"
				TimeSinceChkpt = "00:00:00:00"
			} else {
				if *bt == 0 {
					ct, err := md.GetCreateTime()
					if err != nil {
						return nil, err
					}
					Lag = utils.DataStreamLagTime(*ct)
					TimeSinceChkpt = Lag
				} else {
					Lag = utils.DataStreamLagTime(*bt)
					TimeSinceChkpt = utils.DataStreamLagTime(checkPointLag)
				}

			}

			ib.Set(&InfoAll{Type: info.Groups.DbType, Program: info.Groups.ProcessType, Status: proState, GroupID: info.Groups.GroupID, Lag: Lag, SinceChkpt: TimeSinceChkpt})

		}
	}
	res := ib.ToString()
	return &res, nil
}

func InitRpcClient(log *logrus.Logger) *RpcBody {
	r := new(RpcBody)
	r.Pro = make(map[string]*Processor)
	r.log = log
	return r
}
