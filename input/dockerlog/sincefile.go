package inputdockerlog

import (
	"os"
	"time"
	"unsafe"

	"github.com/tsaikd/KDGoLib/futil"
	"github.com/tsaikd/KDGoLib/mmfile"
)

func NewSinceFile(filepath string) (sincefile *SinceFile, err error) {
	sincefile = &SinceFile{}
	err = sincefile.Open(filepath)
	return
}

type SinceFile struct {
	mmfile mmfile.MMFile
	Since  *time.Time
}

func (t *SinceFile) Open(filepath string) (err error) {
	if err = t.Close(); err != nil {
		return
	}

	if !futil.IsExist(filepath) {
		if err = os.WriteFile(filepath, make([]byte, 32), 0o644); err != nil {
			return
		}
	}

	if t.mmfile, err = mmfile.Open(filepath); err != nil {
		return
	}

	t.Since = (*time.Time)(unsafe.Pointer(&t.mmfile.Data()[0]))

	return
}

func (t *SinceFile) Close() (err error) {
	if t.mmfile != nil {
		if err = t.mmfile.Close(); err != nil {
			return
		}
	}
	t.Since = nil

	return
}
