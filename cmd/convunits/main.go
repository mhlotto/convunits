package main

import (
	"convunits/internal/cli"
	"os"
)

func main() { os.Exit(cli.New(os.Stdout, os.Stderr).Run(os.Args[1:])) }
