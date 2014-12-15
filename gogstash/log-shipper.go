package main

import (
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
