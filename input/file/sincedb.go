package inputfile

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"time"

	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
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

	if err = jsoniter.Unmarshal(raw, &self.SinceDBInfos); err != nil {
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

func (t *InputConfig) CheckSaveSinceDBInfos() (err error) {
	var (
		raw []byte
	)
	if time.Since(t.SinceDBLastSaveTime) > time.Duration(t.SinceDBWriteInterval)*time.Second {
		if raw, err = json.Marshal(t.SinceDBInfos); err != nil {
			log.Errorf("Marshal sincedb failed: %s", err)
			return
		}
		if !bytes.Equal(raw, t.sinceDBLastInfosRaw) {
			err = t.SaveSinceDBInfos()
		}
	}
	return
}

func (t *InputConfig) CheckSaveSinceDBInfosLoop(ctx context.Context) (err error) {
	ticker := time.NewTicker(time.Duration(t.SinceDBWriteInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err = t.CheckSaveSinceDBInfos(); err != nil {
				return
			}
		}
	}
}
