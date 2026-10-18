package main

import (
	"errors"
	"flag"
	"log"
	"os"

	"github.com/MikeFilimonov/go-pwc/config"
	"github.com/MikeFilimonov/go-pwc/core/links"
)

func main() {

	params, err := config.Load(os.Args[1:], os.Stdout)

	if errors.Is(err, flag.ErrHelp) {
		os.Exit(0)
	}

	if err != nil {
		os.Exit(2)
	}

	input, err := links.FixURL(params.InitialURL)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}

	_ = input

	log.Println(params.MaxDepth)

}
