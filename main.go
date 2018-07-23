package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/urfave/cli"
)

var CurrentTime time.Time
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
		cli.IntFlag{
			Name:  "cpu-time-limit",
			Usage: "Limit on CPU time allowed in seconds",
			Value: -1,
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
	elapsedTimeSeconds := time.Now().Sub(CurrentTime).Seconds()
	fmt.Printf("c ================================================================================\n")
	fmt.Printf("c restarts: %12d\n", s.Statistics.RestartCount)
	fmt.Printf("c conflicts: %12d (%.02f / sec)\n", s.Statistics.ConflictCount, float64(s.Statistics.ConflictCount)/elapsedTimeSeconds)
	fmt.Printf("c decisions: %12d (%.02f / sec)\n", s.Statistics.DecisionCount, float64(s.Statistics.DecisionCount)/elapsedTimeSeconds)
	fmt.Printf("c propagations: %12d (%.02f / sec)\n", s.Statistics.PropagationCount, float64(s.Statistics.PropagationCount)/elapsedTimeSeconds)
	fmt.Printf("c reduce DB: %12d\n", s.Statistics.ReduceDBCount)
	fmt.Printf("c removed clause: %12d\n", s.Statistics.RemovedClauseCount)
	fmt.Printf("c cpu time: %12f\n", elapsedTimeSeconds)
}

func setTimeOut(s *Solver, limitTimeSeconds int) {
	if limitTimeSeconds <= 0 {
		return
	}
	go func() {
		<-time.After(time.Duration(limitTimeSeconds) * time.Second)
		fmt.Println("c TIMEOUT")
		if s.Verbosity {
			printStatistics(s)
		}
		fmt.Println("\ns INDETERMINATE")
		os.Exit(0)
	}()
}

func setInterupt(s *Solver) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("c INTERUPT")
		if s.Verbosity {
			printStatistics(s)
		}
		fmt.Println("\ns INDETERMINATE")
		os.Exit(0)
	}()
}

func printModel(s *Solver) {
	fmt.Print("v ")
	for i := 0; i < s.NumVars(); i++ {
		if s.Model[i] == LitBoolTrue {
			fmt.Printf("%d ", i+1)
		} else {
			fmt.Printf("%d ", -(i + 1))
		}
	}
	fmt.Print("0\n")
}

func init() {
	CurrentTime = time.Now()
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
		setTimeOut(solver, c.Int("cpu-time-limit"))
		setInterupt(solver)
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
			printModel(solver)
		} else if status == LitBoolFalse {
			fmt.Println("\ns UNSATISFIABLE")
		}
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
