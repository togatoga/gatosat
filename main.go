package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/k0kubun/pp"
	"github.com/urfave/cli"
)

var DebugMode bool

func GetFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:  "debug,d",
			Usage: "Debug mode",
		},
		cli.StringFlag{
			Name:  "input-file, in",
			Usage: "input cnf file for solving(required)",
			Value: "None",
		},
	}
}

func ValidateFlags(c *cli.Context) (err error) {
	if c.String("input-file") == "None" {
		return fmt.Errorf("input-file is required.")
	}
	return nil
}

func main() {

	app := cli.NewApp()
	app.Name = "gatosat"
	app.Usage = "A CDCL SAT Solver"
	app.Flags = GetFlags()

	app.Before = func(c *cli.Context) error {
		DebugMode = c.Bool("debug")
		return nil
	}

	app.Action = func(c *cli.Context) error {
		var err error
		//validate flag
		err = ValidateFlags(c)
		if err != nil {
			fmt.Println(err)
			cli.ShowAppHelpAndExit(c, 2)
		}

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
