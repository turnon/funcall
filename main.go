package main

import (
	"os"

	"github.com/turnon/funcall/analyzer"
	"github.com/turnon/funcall/graph"
)

func main() {
	args := os.Args[1:len(os.Args)]
	if len(args) == 0 {
		panic("no package given !")
	}

	packageName := args[0]
	result, err := analyzer.Analyze(packageName)
	if err != nil {
		panic(err) // internal error in pointer analysis
	}

	gd := graph.NewGraphData(result)
	gd.Process(args)
	gd.WriteFile("./finalHtml.html")
}
