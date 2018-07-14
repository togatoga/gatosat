package main

type Watcher struct {
	claRef  ClauseReference
	blocker Lit
}

func NewWatcher(cla ClauseReference, p Lit) *Watcher {
	return &Watcher{
		claRef:  cla,
		blocker: p,
	}
}
