package main

import (
	"fmt"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/pointer"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

func main() {
	cfg := &packages.Config{
		Mode:  packages.LoadAllSyntax,
		Tests: false,
		Dir:   "",
	}

	initial, err := packages.Load(cfg, "github.com/turnon/bookmark")
	if err != nil {
		panic(err)
	}

	if packages.PrintErrors(initial) > 0 {
		fmt.Errorf("packages contain errors")
	}

	// Create and build SSA-form program representation.
	prog, pkgs := ssautil.AllPackages(initial, 0)
	prog.Build()

	mains, err := mainPackages(pkgs)
	if err != nil {
		panic(err)
	}

	config := &pointer.Config{
		Mains:          mains,
		BuildCallGraph: true,
	}

	result, err := pointer.Analyze(config)
	if err != nil {
		panic(err) // internal error in pointer analysis
	}

	fmt.Println(len(result.CallGraph.Nodes))

	for fn, _ := range result.CallGraph.Nodes {
		fmt.Println(fn)
	}
}

func mainPackages(pkgs []*ssa.Package) ([]*ssa.Package, error) {
	var mains []*ssa.Package
	for _, p := range pkgs {
		if p != nil && p.Pkg.Name() == "main" && p.Func("main") != nil {
			mains = append(mains, p)
		}
	}
	if len(mains) == 0 {
		return nil, fmt.Errorf("no main packages")
	}
	return mains, nil
}
