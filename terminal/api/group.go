package api

import (
	"fmt"
	"github.com/892294101/dds/dbs/utils"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	_DefaultFilePerm = 0755
)

type ProcessGroupInfo struct {
	GroupId     string
	DBType      string
	ProcessType string
}

func (p *ProcessGroupInfo) InsertGroupInfo(groupName string, dbType string, proType string) error {
	if len(groupName) > 0 && len(dbType) > 0 && len(proType) > 0 {
		p.GroupId = strings.ToUpper(groupName)
		p.DBType = dbType
		p.ProcessType = proType
		return nil
	}
	return errors.Errorf("Cannot be empty")
}

func NewGroupFile() *ProcessGroupInfo {
	return new(ProcessGroupInfo)
}

func (p *ProcessGroupInfo) WriteTo() error {
	home, err := utils.GetHomeDirectory()
	if err != nil {
		return err
	}
	file := filepath.Join(*home, "group", p.GroupId)
	ok := utils.IsFileExist(file)
	if ok {
		return errors.Errorf("group already exists: %v\n", p.GroupId)
	}
	hand, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_SYNC, _DefaultFilePerm)
	if err != nil {
		return errors.Errorf("open file failed the WriteProcessInfo: %v", err)
	}
	var gInfo strings.Builder
	gInfo.WriteString(fmt.Sprintf("%s: %s, %s: %s\n", utils.DBTYPE, p.DBType, utils.PROCESSTYPE, p.ProcessType))
	hand.WriteString(gInfo.String())
	gInfo.Reset()
	hand.Close()
	return nil
}

func (p *ProcessGroupInfo) Remove() error {
	home, err := utils.GetHomeDirectory()
	if err != nil {
		return err
	}
	file := filepath.Join(*home, "group", p.GroupId)
	return os.Remove(file)
}

func (p *ProcessGroupInfo) ReadGroupFileInfo(g string) (*ProcessGroupInfo, error) {
	p.GroupId = strings.ToUpper(g)
	home, err := utils.GetHomeDirectory()
	if err != nil {
		return nil, err
	}
	file := filepath.Join(*home, "group", p.GroupId)
	ok := utils.IsFileExist(file)
	if !ok {
		return nil, errors.Errorf("group not exists: %v", p.GroupId)
	}

	groupIdFile := filepath.Join(*home, "group", p.GroupId)
	r, err := utils.ReadLine(groupIdFile)
	if err != nil {
		return nil, err
	}

	if len(r) == 1 {
		for _, s := range r {
			v := strings.Split(s, ",")
			for _, s2 := range v {
				ind := strings.Index(s2, ":")
				if ind != -1 {
					switch strings.TrimSpace(s2[:ind]) {
					case utils.DBTYPE:
						p.DBType = strings.TrimSpace(s2[ind+1:])
					case utils.PROCESSTYPE:
						p.ProcessType = strings.TrimSpace(s2[ind+1:])
					default:
						return nil, errors.Errorf("missing group file information attribute: %v", p.GroupId)
					}
				} else {
					return nil, errors.Errorf("read group attribute error: %v", p.GroupId)
				}
			}
		}
	} else {
		return nil, errors.Errorf("read group id %v file error", p.GroupId)
	}
	return p, nil
}
