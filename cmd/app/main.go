package main

import "emag-homework/internal/app/bootstrap"

func main() {
	if err := bootstrap.Bootstrap(); err != nil {
		panic(err)
	}
}
