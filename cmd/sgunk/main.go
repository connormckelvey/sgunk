package main

import (
	"log"

	"github.com/connormckelvey/sgunk"
	"github.com/connormckelvey/sgunk/extension/blog"
)

func main() {
	p := sgunk.New(
		sgunk.WithWorkDir("testdata/project1"),
		sgunk.WithExtensions(&blog.Extension{}),
	)

	err := p.Generate()
	if err != nil {
		log.Fatal(err)
	}
}
