package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/pointer"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

//go:embed verd.html
var tmpl string

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

	gd := NewGraphData()

	for _, nodes := range result.CallGraph.Nodes {
		for _, edge := range nodes.Out {
			if strings.Index(edge.String(), "bookmark") <= 0 {
				break
			}
			gd.addLink(edge)
		}
	}

	bytes, _ := json.Marshal(gd)

	toReplace := regexp.MustCompile(`(?s)//start-sub.*//end-sub`)
	finalHtml := toReplace.ReplaceAllLiteralString(tmpl, string(bytes))
	ioutil.WriteFile("./finalHtml.html", []byte(finalHtml), 0644)
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
