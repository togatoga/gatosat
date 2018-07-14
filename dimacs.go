package main

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

func readClause(line string, s *Solver) (lits []Lit, err error) {
	values := strings.Split(line, " ")
	if values[len(values)-1] != "0" {
		return nil, fmt.Errorf("PARSE ERROR! The end of clause is not 0: %s", values[len(values)-1])
	}
	for i := 0; i < len(values)-1; i++ {
		parsedValue, err := strconv.Atoi(values[i])
		if err != nil {
			return nil, err
		}

		value := parsedValue
		neg := false
		if parsedValue < 0 {
			neg = true
			value *= -1
		}
		value--
		if value < 0 {
			return nil, fmt.Errorf("PARSE ERROR! The format of cnf input is worng")
		}

		//TODO
		/* for value >= s.nVars() {
			s.NewValue()
		} */

		lit := NewLit(Var(value), neg)
		lits = append(lits, lit)
	}

	return lits, nil
}

func parseDimacs(in *bufio.Scanner, s *Solver) (err error) {
	vars := 0
	clauses := 0
	cnt := 0
	for in.Scan() {
		line := in.Text()
		line = strings.TrimLeft(line, " ")
		//comment
		if strings.HasPrefix(line, "c") {
			continue
		}
		if strings.HasPrefix(line, "p cnf") {
			values := strings.Split(line, " ")
			vars, err = strconv.Atoi(values[2])
			if err != nil {
				return err
			}
			clauses, err = strconv.Atoi(values[3])
			if err != nil {
				return err
			}
		} else {
			cnt++
			lits, err := readClause(line, s)
			if err != nil {
				return err
			}
			s.addClause(lits)
		}
	}
	if cnt != clauses {
		fmt.Printf("PARSE ERROR! wrong number of clause: %d %d", cnt, clauses)
	}
	_ = vars
	return nil
}
