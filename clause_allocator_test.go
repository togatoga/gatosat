package main

import (
	"math/rand"
	"testing"
)

func BenchmarkNewAllocate(b *testing.B) {
	c := NewClauseAllocator()
	seed := int64(114514)
	rand.Seed(seed)
	for i := 0; i < b.N; i++ {
		size := 100
		clauses := make([]Lit, size)
		for j := 0; j < size; j++ {
			var v Var
			v = Var(j + 1)
			sign := true
			if rand.Int()%2 == 0 {
				sign = false
			}
			clauses[j] = *NewLit(v, sign)
		}
		learnt := true
		if rand.Int()%2 == 0 {
			learnt = false
		}
		c.NewAllocate(clauses, learnt)
	}

}
