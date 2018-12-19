package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSolve(t *testing.T) {
	//SAT
	satDir := "test/sat"
	satFiles, err := ioutil.ReadDir(satDir)
	if err != nil {
		panic(err)
	}
	for _, satFile := range satFiles {
		if satFile.IsDir() {
			continue
		}
		if strings.HasSuffix(satFile.Name(), ".cnf") {
			fileName := filepath.Join(satDir, satFile.Name())

			f, err := os.Open(fileName)
			if err != nil {
				panic(err)
			}
			buf := bufio.NewScanner(f)
			fmt.Println("The solver is solving a sat problem... ", fileName)
			solver := NewSolver()
			err = parseDimacs(buf, solver)
			if err != nil {
				fmt.Println(err, fileName)
				continue
			}
			status := solver.Solve()
			if status != LitBoolTrue {
				err = fmt.Errorf("The solver returns a wrong value for a sat problem: %s", fileName)
				panic(err)
			}
		}
	}
}
