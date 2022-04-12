package main

import (
	"github.com/linkxzhou/prom_merge/merge"
	log "github.com/sirupsen/logrus"
)

func main() {
	if err := merge.NewRootCommand().Execute(); err != nil {
		log.Fatal(err)
	}
}
