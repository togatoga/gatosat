package main

import (
	"fmt"
)

type Solver struct {
	Verbosity    bool
	ClaAllocator *ClauseAllocator
	Clauses      map[ClauseReference]bool
	Watches      map[Lit][]*Watcher
	Assigns      []LiteralBool
	NextVar      Var //Next variable to be created
	VarData      []*VarData
}

func NewSolver() *Solver {
	solver := &Solver{
		Verbosity:    false,
		ClaAllocator: NewClauseAllocator(),
		Clauses:      make(map[ClauseReference]bool),
		Watches:      make(map[Lit][]*Watcher),
		NextVar:      0,
	}
	return solver
}

func (s *Solver) NewVar() Var {
	v := s.NextVar
	s.NextVar++
	s.Assigns = append(s.Assigns, LiteralUndef)
	s.VarData = append(s.VarData, NewVarData(ClaRefUndef, 0))
	return v
}

func (s *Solver) NumVars() int {
	return int(s.NextVar)
}

func (s *Solver) addClause(lits []Lit) (err error) {
	claRef, err := s.ClaAllocator.NewAllocate(lits, false)
	if err != nil {
		return err
	}
	s.Clauses[claRef] = true

	err = s.attachClause(claRef)
	if err != nil {
		return err
	}
	return nil
}

func (s *Solver) attachClause(claRef ClauseReference) (err error) {
	clause, err := s.ClaAllocator.GetClause(claRef)
	if err != nil {
		return err
	}
	if clause.Size() < 2 {
		return fmt.Errorf("The size of clause is less than 2 %v", clause)
	}

	firstLit := clause.At(0)
	secondLit := clause.At(1)

	s.Watches[firstLit.Flip()] = append(s.Watches[firstLit.Flip()], NewWatcher(claRef, firstLit))
	s.Watches[secondLit.Flip()] = append(s.Watches[secondLit.Flip()], NewWatcher(claRef, secondLit))

	return nil
}
