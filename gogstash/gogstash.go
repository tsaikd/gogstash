package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/tsaikd/KDGoLib/env"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/version"
)

var (
	FlDebug = flag.Bool(
		[]string{"-DEBUG"},
		env.GetBool("DEBUG", false),
		"Show DEBUG messages",
	)
	FlConfig = flag.String(
		[]string{"-CONFIG"},
		env.GetString("Config", "config.json"),
		"Path to configuration file",
	)
	FlVersion = flag.Bool(
		[]string{"V", "-VERSION"},
		false,
		"Show version information",
	)
	FlProfile = flag.String(
		[]string{"-PROFILE"},
		env.GetString("PROFILE", ""),
		"Listen http profiling interface, e.g. localhost:6060",
	)
)

func main() {
	flag.Parse()
	mainBody()
}

func mainBody() {
	var (
		conf      config.Config
		err       error
		eventChan = make(chan config.LogEvent, 100)
	)

	if *FlDebug {
		log.SetLevel(log.DebugLevel)
	}

	if *FlVersion {
		version.ShowVersion(os.Stdout)
		return
	}

	if conf, err = config.LoadConfig(*FlConfig); err != nil {
		log.Errorf("Load config failed: %q", *FlConfig)
		return
	}

	if *FlProfile != "" {
		go func() {
			log.Infof("Profile listen http: %q", *FlProfile)
			if err := http.ListenAndServe(*FlProfile, nil); err != nil {
				log.Errorf("Profile listen http failed: %q\n%v", *FlProfile, err)
			}
		}()
	}

	for _, input := range conf.Input() {
		log.Debugf("Init input %q", input.Type())
		if err = input.Event(eventChan); err != nil {
			log.Errorf("input failed: %v", err)
		}
	}

	go func() {
		outputs := conf.Output()
		for {
			select {
			case event := <-eventChan:
				for _, output := range outputs {
					if err = output.Event(event); err != nil {
						log.Errorf("output failed: %v", err)
					}
				}
			}
		}
	}()

	for {
		// infinite sleep
		time.Sleep(1 * time.Hour)
	}
}
