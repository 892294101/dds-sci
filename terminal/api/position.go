package api

import (
	"github.com/892294101/dds/spfile"
	"github.com/892294101/dds/utils"
	"github.com/pkg/errors"
	"regexp"
	"strconv"
	"strings"
)

const (
	ExtTranLogNumPos  = "^((?i)alter)(\\s+)((?i)" + EXTRACT + ")(\\s+)((?:[A-Za-z0-9_]){1,12})(\\s+)((?i)" + TRANLOG + ")(\\s+)((?i)" + LOGNUM + ")(\\s+)(\\d+)(\\s+)((?i)" + LOGPOS + ")(\\s+)(\\d+)$"
	ExtTrailLogSeqRBA = "^((?i)alter)(\\s+)((?i)" + EXTRACT + ")(\\s+)((?:[A-Za-z0-9_]){1,12})(\\s+)((?i)" + TRAILLOG + ")(\\s+)((?i)" + EXTSEQNO + ")(\\s+)(\\d+)(\\s+)((?i)" + EXTRBA + ")(\\s+)(\\d+)$"
	ExtTranLogSCN     = "^((?i)alter)(\\s+)((?i)" + EXTRACT + ")(\\s+)((?:[A-Za-z0-9_]){1,12})(\\s+)((?i)" + SCN + ")(\\s+)(\\d+)$"
)

type TrailAttrBody struct {
	ProcessName string
	sequence    uint64
	rba         uint64
}

func (t *TrailAttrBody) GetTSN() (uint64, uint64) {
	return t.sequence, t.rba
}

func (t *TrailAttrBody) parse(args string, regVal *regexp.Regexp) error {
	matchSet := regVal.FindStringSubmatch(args)
	for i := 0; i < len(matchSet); i++ {
		switch {
		case strings.EqualFold(matchSet[i], EXTRACT):
			t.ProcessName = matchSet[i+2]
			i += 2
		case strings.EqualFold(matchSet[i], EXTSEQNO):
			v, err := strconv.ParseUint(matchSet[i+2], 10, 64)
			if err != nil {
				return errors.Errorf("%v illegal value: %v", EXTSEQNO, matchSet[i+2])
			}
			t.sequence = v
			i += 2
		case strings.EqualFold(matchSet[i], EXTRBA):
			v, err := strconv.ParseUint(matchSet[i+2], 10, 64)
			if err != nil {
				return errors.Errorf("%v illegal value: %v", EXTRBA, matchSet[i+2])
			}
			t.rba = v
			i += 2
		}
	}
	return nil
}

type LSNAttributes struct {
	ProcessName string
	logNumber   uint64
	position    uint64
}

func (m *LSNAttributes) GetLSN() (uint64, uint64) {
	return m.logNumber, m.position
}

func (m *LSNAttributes) parse(args string, regVal *regexp.Regexp, pi *utils.ProcessInfo) error {
	if pi.Groups.DbType == spfile.GetMySQLName() {
		matchSet := regVal.FindStringSubmatch(args)
		for i := 0; i < len(matchSet); i++ {
			switch {
			case strings.EqualFold(matchSet[i], EXTRACT):
				m.ProcessName = matchSet[i+2]
				i += 2
			case strings.EqualFold(matchSet[i], LOGNUM):
				v, err := strconv.ParseUint(matchSet[i+2], 10, 64)
				if err != nil {
					return errors.Errorf("%v illegal value: %v", LOGNUM, matchSet[i+2])
				}
				m.logNumber = v
				i += 2
			case strings.EqualFold(matchSet[i], LOGPOS):
				v, err := strconv.ParseUint(matchSet[i+2], 10, 64)
				if err != nil {
					return errors.Errorf("%v illegal value: %v", LOGPOS, matchSet[i+2])
				}
				m.position = v
				i += 2
			}
		}

	} else {
		return errors.Errorf("syntax is not supported by %v database\n", spfile.GetMySQLName())
	}

	return nil
}

func AttrParse(args []string, pi *utils.ProcessInfo) (interface{}, error) {
	argvs := utils.SliceToString(args, "")
	if argvs != nil {
		etlnp := regexp.MustCompile(ExtTranLogNumPos)
		etlsr := regexp.MustCompile(ExtTrailLogSeqRBA)
		etlscn := regexp.MustCompile(ExtTranLogSCN)
		switch {
		case etlnp.MatchString(*argvs):
			mp := new(LSNAttributes)
			if err := mp.parse(*argvs, etlnp, pi); err != nil {
				return nil, err
			}
			return mp, nil
		case etlsr.MatchString(*argvs):
			mp := new(TrailAttrBody)
			if err := mp.parse(*argvs, etlsr); err != nil {
				return nil, err
			}
			return mp, nil
		case etlscn.MatchString(*argvs):
		default:
			return nil, errors.Errorf("syntax error: %v\n", *argvs)
		}

	}
	return nil, errors.Errorf("unknown error: args %v\n", args)

}
