package version

import (
	"encoding/json"
	"fmt"
	"io"

	log "github.com/Sirupsen/logrus"
)

type Version struct {
	VERSION   string `json:"version"`
	BUILDTIME string `json:"buildtime,omitempty"`
	GITCOMMIT string `json:"gitcommit,omitempty"`
}

var (
	VERSION   = "0.0.7"
	BUILDTIME string
	GITCOMMIT string
)

func ShowVersion(w io.Writer) {
	var (
		raw []byte
		err error
	)

	ver := Version{
		VERSION:   VERSION,
		BUILDTIME: BUILDTIME,
		GITCOMMIT: GITCOMMIT,
	}

	if raw, err = json.MarshalIndent(ver, "", "\t"); err != nil {
		log.Errorf("Marshal failed: %v\n%v", ver, err)
		return
	}
	fmt.Fprintln(w, string(raw))
}
