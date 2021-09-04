package main

import (
	"emag-homework/internal/db/bootstrap"
	"flag"
)

var isNode bool

func init() {
	flag.BoolVar(&isNode, "node", false, "Start a new node")
	flag.Parse()
}

func main() {
	if !isNode {
		if err := bootstrap.StartController(); err != nil {
			panic(err)
		}

		return
	}

	if err := bootstrap.StartNode(); err != nil {
		panic(err)
	}
}
