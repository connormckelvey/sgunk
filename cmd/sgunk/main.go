package main

import (
	"log"
	"os"

	"github.com/connormckelvey/sgunk"
	"github.com/connormckelvey/sgunk/extension/blog"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	p := sgunk.New(
		sgunk.WithWorkDir(wd),
		sgunk.WithExtensions(&blog.Extension{}),
	)

	if err := p.Generate(); err != nil {
		log.Fatal(err)
	}
}
