package inputdockerlog

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/tsaikd/KDGoLib/futil"
)

func NewSinceDB(dbdir string) (sincedb *SinceDB, err error) {
	sincedb = &SinceDB{}
	err = sincedb.Open(dbdir)
	return
}

func MustNewSinceDB(dbdir string) (sincedb *SinceDB) {
	sincedb = &SinceDB{}
	if err := sincedb.Open(dbdir); err != nil {
		panic(err)
	}
	return
}

type SinceDB struct {
	dbdir    string
	SinceMap map[string]*SinceFile
}

func (t *SinceDB) Open(dbdir string) (err error) {
	if err = t.Close(); err != nil {
		return
	}
	t.dbdir = dbdir
	if !futil.IsExist(t.dbdir) {
		if err = os.MkdirAll(t.dbdir, 0755); err != nil {
			return
		}
	}
	fis, err := ioutil.ReadDir(t.dbdir)
	if err != nil {
		return
	}
	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}
		name := fi.Name()
		path := filepath.Join(dbdir, fi.Name())
		sincefile, err := NewSinceFile(path)
		if err != nil {
			return err
		}
		t.SinceMap[name] = sincefile
	}
	return
}

func (t *SinceDB) Close() (err error) {
	for _, sincefile := range t.SinceMap {
		if err2 := sincefile.Close(); err2 != nil {
			if err == nil {
				err = err2
			}
		}
	}
	t.SinceMap = map[string]*SinceFile{}
	return
}

func (t *SinceDB) Get(id string) (since *time.Time, err error) {
	sincefile := t.SinceMap[id]
	if sincefile == nil {
		path := filepath.Join(t.dbdir, id)
		if sincefile, err = NewSinceFile(path); err != nil {
			return
		}
		t.SinceMap[id] = sincefile
	}
	since = sincefile.Since
	return
}

func (t *SinceDB) Del(id string) (err error) {
	sincefile := t.SinceMap[id]
	if sincefile != nil {
		if err = sincefile.Close(); err != nil {
			return
		}
		path := filepath.Join(t.dbdir, id)
		if err = os.Remove(path); err != nil {
			return
		}
	}
	delete(t.SinceMap, id)
	return
}
