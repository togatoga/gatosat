package main

type Var int

const VarUndef Var = -1

type LitBool int

const (
	LitBoolTrue  LitBool = 0
	LitBoolFalse LitBool = 1
	LitBoolUndef LitBool = 2
)

//Lit is a struct for a Literal
type Lit struct {
	X int //A false literal is a odd value(e.g not x1 -> X = 3)
}

const (
	LitUndef = -2
	LitError = -1
)

//NewLit returns a pointer of the Lit
//A false Lit is returned when sign is trues
func NewLit(x Var, sign bool) *Lit {
	var p Lit
	y := 2 * x
	if sign == true {
		y++
	}
	p.X = int(y)
	return &p
}

//Equal a boolean indicating whether p is equal to l
func (l *Lit) Equal(p Lit) bool {
	if l.X != p.X {
		return false
	}
	return true
}

//NotEqual a boolean indicating whether p is NOT equal to l
func (l *Lit) NotEqual(p Lit) bool {
	return !l.Equal(p)
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

func LitToInt(l Lit) int {
	return l.X
}

func (s *Solver) ValueVar(p Var) LitBool {
	return s.Assigns[p]
}

func (s *Solver) ValueLit(p Lit) LitBool {
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
