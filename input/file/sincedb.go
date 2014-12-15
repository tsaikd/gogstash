package inputfile

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/tsaikd/KDGoLib/futil"
)

type SinceDBInfo struct {
	Offset int64 `json:"offset,omitempty"`
}

func (self *InputConfig) LoadSinceDBInfos() (err error) {
	var (
		raw []byte
	)
	log.Debug("LoadSinceDBInfos")
	self.SinceDBInfos = map[string]*SinceDBInfo{}

	if self.SinceDBPath == "" || self.SinceDBPath == "/dev/null" {
		log.Warnf("No valid sincedb path")
		return
	}

	if !futil.IsExist(self.SinceDBPath) {
		log.Debugf("sincedb not found: %q", self.SinceDBPath)
		return
	}

	if raw, err = ioutil.ReadFile(self.SinceDBPath); err != nil {
		log.Errorf("Read sincedb failed: %q\n%s", self.SinceDBPath, err)
		return
	}

	if err = json.Unmarshal(raw, &self.SinceDBInfos); err != nil {
		log.Errorf("Unmarshal sincedb failed: %q\n%s", self.SinceDBPath, err)
		return
	}

	return
}

func (self *InputConfig) SaveSinceDBInfos() (err error) {
	var (
		raw []byte
	)
	log.Debug("SaveSinceDBInfos")
	self.SinceDBLastSaveTime = time.Now()

	if self.SinceDBPath == "" || self.SinceDBPath == "/dev/null" {
		log.Warnf("No valid sincedb path")
		return
	}

	if raw, err = json.Marshal(self.SinceDBInfos); err != nil {
		log.Errorf("Marshal sincedb failed: %s", err)
		return
	}
	self.sinceDBLastInfosRaw = raw

	if err = ioutil.WriteFile(self.SinceDBPath, raw, 0664); err != nil {
		log.Errorf("Write sincedb failed: %q\n%s", self.SinceDBPath, err)
		return
	}

	return
}

func (self *InputConfig) CheckSaveSinceDBInfos() (err error) {
	var (
		raw []byte
	)
	if time.Since(self.SinceDBLastSaveTime) > time.Duration(self.SinceDBWriteInterval)*time.Second {
		if raw, err = json.Marshal(self.SinceDBInfos); err != nil {
			log.Errorf("Marshal sincedb failed: %s", err)
			return
		}
		if bytes.Compare(raw, self.sinceDBLastInfosRaw) != 0 {
			err = self.SaveSinceDBInfos()
		}
	}
	return
}

func (self *InputConfig) CheckSaveSinceDBInfosLoop() (err error) {
	for {
		time.Sleep(time.Duration(self.SinceDBWriteInterval) * time.Second)
		if err = self.CheckSaveSinceDBInfos(); err != nil {
			return
		}
	}
	return
}
