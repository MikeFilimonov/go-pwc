package config

import (
	"flag"
	"fmt"
	"io"
	"strconv"
)

type Config struct {
	MaxDepth int
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

		fmt.Fprint(output, "run -h to get help\n")
		set.PrintDefaults()

	}
	if err := set.Parse(args); err != nil {
		return nil, err
	}

	config := Config{MaxDepth: maxDepth}

	return &config, nil
}
