package main

import (
	"go/parser"
	"log"
	"code.google.com/p/go.tools/importer"
	"code.google.com/p/go.tools/ssa"
	"github.com/arjenroodselaar/hho/hho"
)

func TestEmitExample(t *testing.T) {
	name := "examples/calc.go"
	imp := importer.New(&importer.Config{})

	// Parse the input file.
	f, err := parser.ParseFile(imp.Fset, name, nil, parser.Mode(0))
	if err != nil {
		panic(err)
	}

	imp.CreatePackage(f.Name.Name, f)
	prog := ssa.NewProgram(imp.Fset, ssa.BuilderMode(0))
	if err = prog.CreatePackages(imp); err != nil {
		panic(err)
	}
	prog.BuildAll()

	//pkg := prog.Package(info.Pkg)
	//pkg.DumpTo(os.Stdout)

	//prog.BuildAll()
	//hho.EmitProgram(prog)

	// Create single-file main package and import its dependencies.
	//
	// Create packages for the dependencies.
	//pkg := prog.Package(info.Pkg)
	//pkg.Build()

	//pkg.DumpTo(os.Stdout)

	hho.EmitProgram(prog)
}
