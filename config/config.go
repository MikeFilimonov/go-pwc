package config

import (
	"flag"
	"fmt"
	"io"
	"strconv"
)

type Config struct {
	MaxDepth   int
	InitialURL string
}

const defaultMaxDepth int = 5

func Load(args []string, output io.Writer) (*Config, error) {

	set := flag.NewFlagSet("pwc", flag.ContinueOnError)

	usage := "parallel web crawler max depth (5, by default)"
	maxDepth := defaultMaxDepth

	set.SetOutput(output)

	set.Func("maxDepth", usage, func(value string) error {

		result, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		if result <= 0 {
			return fmt.Errorf("please, use a positive number instead")
		}
		maxDepth = result

		return nil

	})

	set.Usage = func() {

		fmt.Fprintln(output, "run -h to get help")
		set.PrintDefaults()

	}
	if err := set.Parse(args); err != nil {
		return nil, err
	}

	// no url
	if set.NArg() == 0 {

		description := "please, provide the initial URL to start processing"
		fmt.Fprintln(output, description)
		set.Usage()
		return nil, fmt.Errorf("%s", description)
	}

	config := Config{
		MaxDepth:   maxDepth,
		InitialURL: set.Arg(0),
	}

	return &config, nil
}
