package main

import (
	"fmt"
	"math"
)

type ClauseReference uint32

const ClaRefUndef ClauseReference = math.MaxUint32

//ClauseAllocator is a allocator for the clause
//NOTE we need to improve the performance of alloc/free in the future
type ClauseAllocator struct {
	Qhead   ClauseReference             //the head of the ClauseAllocator
	Clauses map[ClauseReference]*Clause // the performace of the map is really bad. we should replace it with the array?
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

//FreeClause deletes the clause if the clause is allocated
func (c *ClauseAllocator) FreeClause(claRef ClauseReference) {
	if _, ok := c.Clauses[claRef]; ok {
		delete(c.Clauses, claRef)
		return
	}
	panic(fmt.Errorf("The clause is not allocated: %d", claRef))
}
