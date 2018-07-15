package main

import (
	"fmt"
)

type Solver struct {
	Verbosity    bool
	ClaAllocator *ClauseAllocator         //The allocator for clause
	Clauses      map[ClauseReference]bool //List of problem clauses.
	Watches      map[Lit][]*Watcher       //'watches[lit]' is a list of constraints watching 'lit' (will go there if literal becomes true).
	Assigns      []LiteralBool            //The current assignments.
	Qhead        int                      // Head of queue (as index into the trail -- no more explicit propagation queue in MiniSat).
	Trail        []Lit                    //Assignment stack; stores all assigments made in the order the were made.
	TrailLim     []int                    //Separator indices for different decision levels in 'trail'.
	NextVar      Var                      //Next variable to be created.
	VarData      []*VarData               //Stores reason and level for each variable.
}

func NewSolver() *Solver {
	solver := &Solver{
		Verbosity:    false,
		ClaAllocator: NewClauseAllocator(),
		Clauses:      make(map[ClauseReference]bool),
		Watches:      make(map[Lit][]*Watcher),
		Qhead:        0,
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

func (s *Solver) Value(p Lit) LiteralBool {
	if s.Assigns[p.Var()] == LiteralUndef {
		return LiteralUndef
	} else if s.Assigns[p.Var()] == LiterealTrue {
		if !p.Sign() {
			return LiterealTrue
		}
	} else if s.Assigns[p.Var()] == LiteralFalse {
		if p.Sign() {
			return LiterealTrue
		}
	}
	return LiteralFalse
}

func (s *Solver) NumVars() int {
	return int(s.NextVar)
}

func (s *Solver) UncheckedEnqueue(p Lit, from ClauseReference) {
	if DebugMode {
		if s.Value(p) != LiteralUndef {
			panic(fmt.Sprintf("The assign is not LiteralUndef: Value(%d) = %v", p, s.Value(p)))
		}
	}

	if !p.Sign() {
		s.Assigns[p.Var()] = LiterealTrue
	} else {
		s.Assigns[p.Var()] = LiteralFalse
	}
	s.VarData[p.Var()] = NewVarData(from, s.DecisionLevel())
	s.Trail = append(s.Trail, p)
}

func (s *Solver) DecisionLevel() int {
	return len(s.TrailLim)
}

func (s *Solver) addClause(lits []Lit) (err error) {

	if len(lits) == 1 {
		panic("TODO")
	} else {
		claRef, err := s.ClaAllocator.NewAllocate(lits, false)
		if err != nil {
			return err
		}
		s.Clauses[claRef] = true

		err = s.attachClause(claRef)
		if err != nil {
			return err
		}
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
