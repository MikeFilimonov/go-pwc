package main

import (
	"log"
	"os"

	"github.com/MikeFilimonov/go-pwc/config"
)

func main() {

	params, err := config.Load(os.Args[1:])
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Println(params.MaxDepth)

}
