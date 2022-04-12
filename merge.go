package main

import (
	"github.com/linkxzhou/prom_merge/cmd"
	log "github.com/sirupsen/logrus"
)

func main() {
	if err := cmd.NewRootCommand().Execute(); err != nil {
		log.Fatal(err)
	}
}
