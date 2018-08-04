package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const (
	// SATEXITCODE is the exit code for UNKNOWN
	SATEXITCODE = 10
	//UNSATEXITCODE is the exit code for UNKNOWN
	UNSATEXITCODE = 20
	// UNKNOWNEXITCODE is the exit code for UNKNOWN
	UNKNOWNEXITCODE = 0
)

var CurrentTime time.Time

var (
	DebugMode    = kingpin.Flag("debug", "Debug mode").Short('d').Bool()
	Verbose      = kingpin.Flag("verbose", "Vervosity mode").Short('v').Default("true").Bool()
	InputFile    = kingpin.Arg("input-file", "Input cnf file for solving").Required().File()
	OutputFile   = kingpin.Arg("output-file", "Output result file").String()
	CPUTimeLimit = kingpin.Flag("cpu-time-limit", "Limit on CPU time allowed in seconds").Int()
)

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

func writeOutputFile(file string, s *Solver, status LitBool) error {
	var fp *os.File

	if _, err := os.Stat(file); os.IsNotExist(err) {
		fp, err = os.Create(file)
		if err != nil {
			return err
		}
	} else {
		fp, err = os.Create(file)
		if err != nil {
			return err
		}
	}
	defer fp.Close()

	if status == LitBoolTrue {
		for i := 0; i < s.NumVars(); i++ {
			if s.Model[i] == LitBoolTrue {
				fp.WriteString(fmt.Sprintf("%d ", i+1))
			} else {
				fp.WriteString(fmt.Sprintf("%d ", -(i + 1)))
			}
		}
		fp.WriteString("0\n")
	} else if status == LitBoolFalse {
		fp.WriteString("UNSAT")
	}
	return nil
}

func init() {
	CurrentTime = time.Now()
}

func run() int {
	//input
	inFp := *InputFile
	defer inFp.Close()
	in := bufio.NewScanner(inFp)

	solver := NewSolver()
	setTimeOut(solver, *CPUTimeLimit)
	setInterupt(solver)

	err := parseDimacs(in, solver)
	if err != nil {
		return UNKNOWNEXITCODE 
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

	if OutputFile != nil {
		writeOutputFile(*OutputFile, solver, status)
	}
	if status == LitBoolTrue {
		return SATEXITCODE
	} else if status == LitBoolFalse {
		return UNSATEXITCODE
	}

	return UNKNOWNEXITCODE
}

func main() {
	kingpin.Version("0.0.1")
	kingpin.Parse()
	os.Exit(run())
}
