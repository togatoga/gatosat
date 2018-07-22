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

func (w *Watcher) Equal(wr Watcher) bool {
	if w.claRef == wr.claRef {
		return true
	}
	return false
}

func RemoveWatcher(watches map[Lit][]*Watcher, x Lit, watcher Watcher) {
	startCopyIdx := -1
	//Find the index of watcher
	for i := 0; i < len(watches[x]); i++ {
		if watches[x][i].Equal(watcher) {
			startCopyIdx = i
			break
		}
	}
	//Copy the rest of watcher exclude the value of startCopyIdx
	for copiedIdx := startCopyIdx; copiedIdx < len(watches[x])-1; copiedIdx++ {
		watches[x][copiedIdx] = watches[x][copiedIdx+1]
	}
	//pop
	watches[x] = watches[x][:len(watches[x])-1]
}
