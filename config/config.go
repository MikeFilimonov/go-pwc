package config

import "flag"

type Config struct {
	MaxDepth int
}

const defaultMaxDepth int = 5

func Load(args []string) (*Config, error) {

	set := flag.NewFlagSet("pwc", flag.ExitOnError)
	maxDepth := set.Int("maxDepth", defaultMaxDepth, "parallel web crawler max depth")

	if err := set.Parse(args); err != nil {
		return nil, err
	}

	config := Config{MaxDepth: *maxDepth}

	return &config, nil
}
