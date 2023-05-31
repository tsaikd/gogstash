package inputexec

import (
	"database/sql/driver"

	"github.com/tsaikd/KDGoLib/enumutil"
)

type MsgType int8

const (
	MsgTypeText MsgType = 1 + iota
	MsgTypeJson
)

var msgTypeEnum = enumutil.NewEnumFactory().
	Add(MsgTypeText, "text").
	Add(MsgTypeJson, "json").
	Build()

func (t MsgType) String() string {
	return msgTypeEnum.String(t)
}

func (t MsgType) MarshalJSON() ([]byte, error) {
	return msgTypeEnum.MarshalJSON(t)
}

func (t *MsgType) UnmarshalJSON(b []byte) (err error) {
	return msgTypeEnum.UnmarshalJSON(t, b)
}

func (t *MsgType) Scan(value any) (err error) {
	return msgTypeEnum.Scan(t, value)
}

func (t MsgType) Value() (v driver.Value, err error) {
	return msgTypeEnum.Value(t)
}

func IsMsgType(s string) bool {
	return msgTypeEnum.IsEnumString(s)
}

func ParseMsgType(s string) MsgType {
	enum, err := msgTypeEnum.ParseString(s)
	if err != nil {
		return 0
	}
	return enum.(MsgType)
}
