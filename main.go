package main

import (
	"errors"
	"flag"
	"log"
	"os"

	"github.com/MikeFilimonov/go-pwc/config"
)

func main() {

	params, err := config.Load(os.Args[1:], os.Stdout)

	if errors.Is(err, flag.ErrHelp) {
		os.Exit(0)
	}

	if err != nil {
		os.Exit(2)
	}

	log.Println(params.MaxDepth)

}
