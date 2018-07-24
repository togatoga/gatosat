# gatosat
A gatosat is a cdcl SAT solver written in golang and inspired by [minisat](https://github.com/niklasso/minisat). Most parts of the algorithm and data structures are based on minisat.

**Under Construction Project**

## How to install
```bash
go get github.com/togatoga/gatosat && go install github.com/togatoga/gatosat
```

## How to use
### Solving SAT Problem(.cnf)

```bash
# usage: gatosat [<flags>] <input-file> [<output-file>]  

# solve problem.cnf
gatosat problem.cnf
# solve problem.cnf and write the output into output.txt
gatosat problem.cnf output.txt
```

`gatosat --help` shows more useful options. Please check it.


## Algorithm
- CDCL
- VSIDS
- Luby Restart
- Two Literal watching
