package inputfile

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"time"

	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"github.com/tsaikd/KDGoLib/futil"
)

const devNull = "/dev/null"

type SinceDBInfo struct {
	Offset int64 `json:"offset,omitempty"`
}

func (t *InputConfig) LoadSinceDBInfos() (err error) {
	var (
		raw []byte
	)
	log.Debug("LoadSinceDBInfos")
	t.SinceDBInfos = map[string]*SinceDBInfo{}

	if t.SinceDBPath == "" || t.SinceDBPath == devNull {
		log.Warnf("No valid sincedb path")
		return
	}

	if !futil.IsExist(t.SinceDBPath) {
		log.Debugf("sincedb not found: %q", t.SinceDBPath)
		return
	}

	if raw, err = os.ReadFile(t.SinceDBPath); err != nil {
		log.Errorf("Read sincedb failed: %q\n%s", t.SinceDBPath, err)
		return
	}

	if err = jsoniter.Unmarshal(raw, &t.SinceDBInfos); err != nil {
		log.Errorf("Unmarshal sincedb failed: %q\n%s", t.SinceDBPath, err)
		return
	}

	return
}

func (t *InputConfig) SaveSinceDBInfos() (err error) {
	var (
		raw []byte
	)
	log.Debug("SaveSinceDBInfos")
	t.SinceDBLastSaveTime = time.Now()

	if t.SinceDBPath == "" || t.SinceDBPath == devNull {
		log.Warnf("No valid sincedb path")
		return
	}

	if raw, err = json.Marshal(t.SinceDBInfos); err != nil {
		log.Errorf("Marshal sincedb failed: %s", err)
		return
	}
	t.sinceDBLastInfosRaw = raw

	if err = os.WriteFile(t.SinceDBPath, raw, 0o664); err != nil {
		log.Errorf("Write sincedb failed: %q\n%s", t.SinceDBPath, err)
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
