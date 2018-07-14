package main

import (
	"bufio"
	"log"
	"os"

	"github.com/k0kubun/pp"
	"github.com/urfave/cli"
)

var DebugMode bool

func main() {

	app := cli.NewApp()
	app.Name = "gatosat"
	app.Usage = "A CDCL SAT Solver"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug,d",
			Usage: "Debug mode",
		},
		cli.StringFlag{
			Name:  "input-file, in",
			Usage: "input cnf file for solving",
		},
	}

	app.Before = func(c *cli.Context) error {
		DebugMode = c.Bool("debug")
		return nil
	}

	app.Action = func(c *cli.Context) error {
		//input
		inputFile := c.String("input-file")
		fp, err := os.Open(inputFile)
		if err != nil {
			return err
		}
		in := bufio.NewScanner(fp)
		solver := NewSolver()
		err = parseDimacs(in, solver)
		if err != nil {
			return err
		}
		pp.Println(solver)

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
