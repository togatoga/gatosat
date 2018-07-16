package main

type Var int

const VarUndef Var = -1

type LitBool int

const (
	LitBoolTrue  LitBool = 0
	LitBoolFalse LitBool = 1
	LitBoolUndef LitBool = 2
)

type Lit struct {
	X int
}

const (
	LitUndef = -2
	LitError = -1
)

func NewLit(x Var, sign bool) *Lit {
	var p Lit
	y := 2 * x
	if sign == true {
		y++
	}
	p.X = int(y)
	return &p
}

func (l *Lit) Equal(p Lit) bool {
	if l.X != p.X {
		return false
	}
	return true
}

func (l *Lit) NotEqual(p Lit) bool {
	if l.X == p.X {
		return false
	}
	return true
}

func (l *Lit) Less(p Lit) bool {
	if l.X > p.X {
		return false
	}
	return true
}

func (l *Lit) Sign() bool {
	if l.X&1 == 0 {
		return false
	}
	return true
}

func (l *Lit) Flip() Lit {
	x := l.Var()
	return *NewLit(x, !l.Sign())
}

func (l *Lit) Var() Var {
	return Var(l.X >> 1)
}

func (l *Lit) ToInt() int {
	return l.X
}
