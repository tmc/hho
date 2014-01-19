hho
===

A Golang to HHAS (HHVM (HipHop Virtual Machine) Assembler) compiler

Testing
=======
```sh
go get github.com/arjenroodselaar/hho
go test github.com/arjenroodselaar/hho
```

Running HHVM as 'as'
====================
./hhvm -v Eval.AllowHhas=true <file>

Plan of Attack
==============
Using ssadump as an example, we're going to do the same thing.
But instead of spitting out SSA, we're pooping out HHAS.

Bitchin'

References
==========

- https://github.com/facebook/hhvm/blob/master/hphp/doc/bytecode.specification
- https://github.com/facebook/hhvm/blob/master/hphp/runtime/vm/as.cpp
