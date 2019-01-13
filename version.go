package main

import "github.com/tsaikd/KDGoLib/version"

var (
	// Version for gogstash
	Version = "0.1.18-dev"
)

func init() {
	version.VERSION = Version
}
