package main

import (
	"fmt"
	"math"
)

//ClauseAllocator
type ClauseReference uint32

const ClaRefUndef ClauseReference = math.MaxUint32

type ClauseAllocator struct {
	Qhead   ClauseReference //Allocate
	Clauses map[ClauseReference]*Clause
}

func NewClauseAllocator() *ClauseAllocator {
	return &ClauseAllocator{Qhead: 0, Clauses: make(map[ClauseReference]*Clause)}
}

func (c *ClauseAllocator) NewAllocate(lits []Lit, learnt bool) (ClauseReference, error) {
	cref := c.Qhead
	c.Clauses[cref] = NewClause(lits, false, learnt)
	c.Qhead++
	return cref, nil
}

func (c *ClauseAllocator) GetClause(claRef ClauseReference) (clause *Clause) {
	if clause, ok := c.Clauses[claRef]; ok {
		return clause
	}
	panic(fmt.Errorf("The clause is not allocated: %d", claRef))
}

const (
	ExistMark   uint = iota
	DeletedMark uint = iota
)

//Clause
type Header struct {
	Mark     uint
	Learnt   bool
	HasExtra bool
	Size     int
}

type Clause struct {
	header Header
	Data   []Lit
	Act    float32
}

func NewClause(ps []Lit, useExtra, learnt bool) *Clause {
	var c Clause
	c.header.Mark = ExistMark
	c.header.Learnt = learnt
	c.header.HasExtra = useExtra
	c.header.Size = len(ps)

	c.Data = make([]Lit, len(ps))
	copy(c.Data, ps)

	c.Act = 0

	return &c
}

func (c *Clause) Size() int {
	return c.header.Size
}

func (c *Clause) Learnt() bool {
	return c.header.Learnt
}

func (c *Clause) HasExtra() bool {
	return c.header.HasExtra
}

func (c *Clause) SetMark(mark uint) {
	c.header.Mark = mark
}

func (c *Clause) Mark() uint {
	return c.header.Mark
}

func (c *Clause) At(i int) Lit {
	return c.Data[i]
}

func (c *Clause) Pop() {
	if c.Size() == 0 {
		panic(fmt.Errorf("Pop empty clause"))
	}
	c.header.Size -= 1
}

func (c *Clause) Last() Lit {
	return c.Data[c.Size()-1]
}

func (c *Clause) Activity() float32 {
	return c.Act
}

func (s *Solver) removeSatisfied(data *[]ClauseReference) {
	copiedIdx := 0

	for lastIdx := 0; lastIdx < len(*data); lastIdx++ {
		c := s.ClaAllocator.GetClause((*data)[lastIdx])
		if s.satisfied(c) {
			s.removeClause((*data)[lastIdx])
		} else {
			//Trim Clause
			if !(s.ValueLit(c.At(0)) == LitBoolUndef && s.ValueLit(c.At(1)) == LitBoolUndef) {
				panic(fmt.Errorf("The 0th and 1th of clause value is not LitBoolUndef: v1: %d = %d v2: %d = %d", c.At(0), s.ValueLit(c.At(0)), c.At(1), s.ValueLit(c.At(1))))
			}
			for k := 2; k < c.Size(); k++ {
				if s.ValueLit(c.At(k)) == LitBoolFalse {
					c.Data[k] = c.Last()
					k--
					c.Pop()
				}
			}
			(*data)[copiedIdx] = (*data)[lastIdx]
			copiedIdx++
		}
	}
	(*data) = (*data)[:copiedIdx]
}

func (s *Solver) detachClause(cr ClauseReference) {
	c := s.ClaAllocator.GetClause(cr)
	if c.Size() <= 1 {
		panic(fmt.Errorf("The size of clause is less than 2: %d", c.Size()))
	}
	firstLit := c.At(0)
	secondLit := c.At(1)
	RemoveWatcher(s.Watches, firstLit.Flip(), NewWatcher(cr, secondLit))
	RemoveWatcher(s.Watches, secondLit.Flip(), NewWatcher(cr, firstLit))
	if c.Learnt() {
		s.Statistics.NumLearnts--
	} else {
		s.Statistics.NumClauses--
	}
}

func (s *Solver) locked(c *Clause) bool {
	firstLit := c.At(0)
	if s.ValueLit(firstLit) == LitBoolTrue && s.Reason(firstLit.Var()) != ClaRefUndef {
		return true
	}
	return false
}

func (s *Solver) satisfied(c *Clause) bool {
	for i := 0; i < c.Size(); i++ {
		if s.ValueLit(c.At(i)) == LitBoolTrue {
			return true
		}
	}
	return false
}

func (s *Solver) removeClause(cr ClauseReference) {
	c := s.ClaAllocator.GetClause(cr)
	s.detachClause(cr)
	firstLit := c.At(0)
	if s.locked(c) {
		s.VarData[firstLit.Var()].Reason = ClaRefUndef
	}
	c.SetMark(DeletedMark)
	delete(s.ClaAllocator.Clauses, cr)
}

func (s *Solver) attachClause(claRef ClauseReference) (err error) {
	clause := s.ClaAllocator.GetClause(claRef)

	if clause.Size() < 2 {
		return fmt.Errorf("The size of clause is less than 2 %v", clause)
	}

	firstLit := clause.At(0)
	secondLit := clause.At(1)
	s.Watches.Append(firstLit.Flip(), NewWatcher(claRef, secondLit))
	s.Watches.Append(secondLit.Flip(), NewWatcher(claRef, firstLit))

	if clause.Learnt() {
		s.Statistics.NumLearnts++
	} else {
		s.Statistics.NumClauses++
	}
	return nil
}
