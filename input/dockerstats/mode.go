package inputdockerstats

import "github.com/tsaikd/KDGoLib/enumutil"

// Mode enum for docker stats log mode
type Mode int8

// List all valid enum of Mode
const (
	ModeFull Mode = 1 + iota
	ModeSimple
)

var factoryMode = enumutil.NewEnumFactory().
	Add(ModeFull, "full").
	Add(ModeSimple, "simple").
	Build()

func (t Mode) String() string {
	return factoryMode.String(t)
}

// MarshalJSON return jsonfy []byte of enum
func (t Mode) MarshalJSON() ([]byte, error) {
	return factoryMode.MarshalJSON(t)
}

// UnmarshalJSON decode json data to enum
func (t *Mode) UnmarshalJSON(b []byte) (err error) {
	return factoryMode.UnmarshalJSON(t, b)
}

// IsMode check string is valid enum
func IsMode(s string) bool {
	return factoryMode.IsEnumString(s)
}

// ParseMode string to enum
func ParseMode(s string) Mode {
	enum, err := factoryMode.ParseString(s)
	if err != nil {
		return 0
	}
	return enum.(Mode)
}
