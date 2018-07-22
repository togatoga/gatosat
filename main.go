package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli"
)

var DebugMode bool

func GetFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:  "debug,d",
			Usage: "Debug mode",
		},

		cli.BoolTFlag{
			Name:  "verbosity,verb",
			Usage: "Verbosity mode",
		},
		cli.StringFlag{
			Name:  "input-file, in",
			Usage: "Input cnf file for solving(required)",
			Value: "None",
		},

		cli.StringFlag{
			Name:  "result-output-file, out",
			Usage: "Output file",
		},
	}
}

func ValidateFlags(c *cli.Context) (err error) {
	if c.String("input-file") == "None" {
		return fmt.Errorf("input-file is required.")
	}
	return nil
}

func printProblemStatistics(s *Solver) {
	fmt.Printf("c ============================[ Problem Statistics ]=============================\n")
	fmt.Printf("c |                                                                             |\n")
	fmt.Printf("c |  Number of variables:  %12d                                         |\n", s.NumVars())
	fmt.Printf("c |  Number of clauses:    %12d                                         |\n", s.NumClauses())
	fmt.Printf("c ================================================================================\n")
}

func printStatistics(s *Solver) {
	fmt.Printf("c ================================================================================\n")
	fmt.Printf("c restarts: %12d\n", s.Statistics.RestartCount)
	fmt.Printf("c conflicts: %12d\n", s.Statistics.ConflictCount)
	fmt.Printf("c decisions: %12d\n", s.Statistics.DecisionCount)
	fmt.Printf("c propagations: %12d\n", s.Statistics.PropagationCount)
	fmt.Printf("c reduce DB: %12d\n", s.Statistics.ReduceDBCount)
	fmt.Printf("c removed clause: %12d\n", s.Statistics.RemovedClauseCount)
}

func SetInterupt(s *Solver) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		if s.Verbosity {
			printStatistics(s)
		}
		fmt.Println("\ns INDETERMINATE")
		os.Exit(0)
	}()
}

func main() {

	app := cli.NewApp()
	app.Name = "gatosat"
	app.Usage = "A CDCL SAT Solver written in Go"
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
		defer fp.Close()
		if err != nil {
			return err
		}
		in := bufio.NewScanner(fp)
		solver := NewSolver(c)
		SetInterupt(solver)
		err = parseDimacs(in, solver)
		if err != nil {
			return err
		}
		if solver.Verbosity {
			printProblemStatistics(solver)
		}
		status := solver.Solve()

		if solver.Verbosity {
			printStatistics(solver)
		}
		if status == LitBoolTrue {
			fmt.Println("\ns SATISFIABLE")
		} else if status == LitBoolFalse {
			fmt.Println("\ns UNSATISIABLE")
		}
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
