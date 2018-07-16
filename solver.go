package main

import (
	"fmt"
)

type Solver struct {
	Verbosity    bool
	ClaAllocator *ClauseAllocator         //The allocator for clause
	Clauses      map[ClauseReference]bool //List of problem clauses.
	Watches      map[Lit][]*Watcher       //'watches[lit]' is a list of constraints watching 'lit' (will go there if literal becomes true).
	Assigns      []LitBool                //The current assignments.
	Qhead        int                      //Head of queue (as index into the trail -- no more explicit propagation queue in MiniSat).
	Trail        []Lit                    //Assignment stack; stores all assigments made in the order the were made.
	TrailLim     []int                    //Separator indices for different decision levels in 'trail'.
	NextVar      Var                      //Next variable to be created.
	VarData      []*VarData               //Stores reason and level for each variable.
	OK           bool                     //If FALSE, the constraints are already unsatisfiable. No part of the solver state may be used!
	Seen         []bool                   //The seen variable for clause learning
}

func NewSolver() *Solver {
	solver := &Solver{
		Verbosity:    false,
		ClaAllocator: NewClauseAllocator(),
		Clauses:      make(map[ClauseReference]bool),
		Watches:      make(map[Lit][]*Watcher),
		Qhead:        0,
		NextVar:      0,
		OK:           true,
	}
	return solver
}

func (s *Solver) NewVar() Var {
	v := s.NextVar
	s.NextVar++
	s.Assigns = append(s.Assigns, LitBoolUndef)
	s.VarData = append(s.VarData, NewVarData(ClaRefUndef, 0))
	s.Seen = append(s.Seen, false)
	return v
}

func (s *Solver) Value(p Lit) LitBool {
	if s.Assigns[p.Var()] == LitBoolUndef {
		return LitBoolUndef
	} else if s.Assigns[p.Var()] == LitBoolTrue {
		if !p.Sign() {
			return LitBoolTrue
		}
	} else if s.Assigns[p.Var()] == LitBoolFalse {
		if p.Sign() {
			return LitBoolTrue
		}
	}
	return LitBoolFalse
}

func (s *Solver) NumVars() int {
	return int(s.NextVar)
}

func (s *Solver) UncheckedEnqueue(p Lit, from ClauseReference) {
	if DebugMode {
		if s.Value(p) != LitBoolUndef {
			panic(fmt.Sprintf("The assign is not LiteralUndef: Value(%d) = %v", p, s.Value(p)))
		}
	}

	if !p.Sign() {
		s.Assigns[p.Var()] = LitBoolTrue
	} else {
		s.Assigns[p.Var()] = LitBoolFalse
	}
	s.VarData[p.Var()] = NewVarData(from, s.DecisionLevel())
	s.Trail = append(s.Trail, p)
}

func (s *Solver) Propagate() ClauseReference {
	confl := ClaRefUndef

	for s.Qhead < len(s.Trail) {
		p := s.Trail[s.Qhead]
		s.Qhead++
		lastIdx := 0
		copiedIdx := 0
		for lastIdx < len(s.Watches[p]) {
			watcher := s.Watches[p][lastIdx]
			blocker := watcher.blocker
			// Try to avoid inspecting the clause.
			if s.Value(blocker) == LitBoolTrue {
				s.Watches[p][copiedIdx] = s.Watches[p][lastIdx]
				lastIdx++
				copiedIdx++
				continue
			}

			// Make sure the false literal is data[1]
			cr := watcher.claRef
			clause, err := s.ClaAllocator.GetClause(cr)
			if err != nil {
				panic(err)
			}
			falseLit := p.Flip()
			if clause.At(0) == falseLit {
				clause.Data[0], clause.Data[1] = clause.Data[1], falseLit
			}
			if clause.At(1) != falseLit {
				panic(fmt.Errorf("The 1th literal is not falseLit: %v %v", clause.At(1), falseLit))
			}
			lastIdx++

			// If 0th watch is true, then clause is already satisfied
			firstLiteral := clause.At(0)
			w := NewWatcher(cr, firstLiteral)
			if firstLiteral != blocker && s.Value(firstLiteral) == LitBoolTrue {
				s.Watches[p][copiedIdx] = w
				copiedIdx++
				continue
			}

			// Look for new watch:
			for i := 2; i < clause.Size(); i++ {
				//Find the candidate for watching
				if s.Value(clause.At(i)) != LitBoolFalse {
					clause.Data[1], clause.Data[i] = clause.Data[i], falseLit
					x := clause.At(1)
					s.Watches[x.Flip()] = append(s.Watches[x.Flip()], w)
					goto NextClause
				}
			}
			// Did not find watch -- clause is unit under assignment:
			s.Watches[p][copiedIdx] = w
			copiedIdx++
			if s.Value(firstLiteral) == LitBoolFalse {
				confl = cr
				s.Qhead = len(s.Trail)
				//Copy the remaining watches:
				for lastIdx < len(s.Watches[p]) {
					s.Watches[p][copiedIdx] = s.Watches[p][lastIdx]
					lastIdx++
					copiedIdx++
				}
			} else {
				s.UncheckedEnqueue(firstLiteral, cr)
			}
		NextClause:
		}
		s.Watches[p] = s.Watches[p][:copiedIdx]
	}

	return confl
}

func (s *Solver) DecisionLevel() int {
	return len(s.TrailLim)
}

func (s *Solver) addClause(lits []Lit) bool {
	// What clause is empty means that the problem is unsatisfiable
	if len(lits) == 0 {
		s.OK = false
	} else if len(lits) == 1 {
		s.UncheckedEnqueue(lits[0], ClaRefUndef)
		confl := s.Propagate()
		//Found conflict
		if confl != ClaRefUndef {
			s.OK = false
		}
	} else {
		claRef, err := s.ClaAllocator.NewAllocate(lits, false)
		if err != nil {
			panic(err)
		}
		s.Clauses[claRef] = true
		err = s.attachClause(claRef)
		if err != nil {
			panic(err)
		}
	}
	return true
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

func (s *Solver) Reason(x Var) ClauseReference {
	return s.VarData[x].Reason
}

func (s *Solver) Level(x Var) int {
	return s.VarData[x].Level
}

func (s *Solver) Analyze(confl ClauseReference) (learntClause []Lit, backTrackLevel int) {
	var p Lit
	p.X = LitUndef

	pathConflict := 0
	idx := len(s.Trail) - 1

	learntClause = append(learntClause, p) // (leave room for the asserting literal)
	for {
		if confl == ClaRefUndef {
			panic("The conflict doesn't point any regisions")
		}
		conflCla, err := s.ClaAllocator.GetClause(confl)
		if err != nil {
			panic(err)
		}
		var startIndex int
		if p.X == LitUndef {
			startIndex = 0
		} else {
			startIndex = 1
		}
		for i := startIndex; i < conflCla.Size(); i++ {
			q := conflCla.At(i)
			if !s.Seen[q.Var()] && s.Level(q.Var()) > 0 {
				//TODO
				s.Seen[q.Var()] = true
				if s.Level(q.Var()) >= s.DecisionLevel() {
					pathConflict++
				} else {
					learntClause = append(learntClause, q)
				}
			}
		}
		// Select next clause to look at:
		update := true
		for update {
			p = s.Trail[idx]
			update = !s.Seen[p.Var()]
			idx--
		}

		confl = s.Reason(p.Var())
		s.Seen[p.Var()] = false
		pathConflict--
		if pathConflict <= 0 {
			break
		}
	}
	learntClause[0] = p.Flip()

	//TODO
	//Simplify conflict clause:

	analyzeToClear := make([]Lit, len(learntClause))
	copy(analyzeToClear, learntClause)

	if len(learntClause) == 1 {
		backTrackLevel = 0
	} else {
		maxIdx := 1
		// Find the first literal assigned at the next-highest level:
		for i := 2; i < len(learntClause); i++ {
			if s.Level(learntClause[i].Var()) > s.Level(learntClause[maxIdx].Var()) {
				maxIdx = i
			}
		}

		learntClause[maxIdx], learntClause[1] = learntClause[1], learntClause[maxIdx]
		backTrackLevel = s.Level(learntClause[maxIdx].Var())
	}
	for _, lit := range analyzeToClear {
		s.Seen[lit.Var()] = false
	}

	return learntClause, backTrackLevel
}

func (s *Solver) Search() LitBool {
	if !s.OK {
		panic("s.OK is false")
	}

	for {
		confl := s.Propagate()

		if confl != ClaRefUndef {
			//Conflict

			//If the decision level is 0, the problem is unsatisfiable.
			if s.DecisionLevel() == 0 {
				return LitBoolFalse
			}

			/* learntClause := []Lit{}
			backTrackLevel := math.MaxInt32 */
			//TODO
			//Analyze

		}
	}

	return LitUndef
}
