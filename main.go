package main

import (
	"github.com/turnon/funcall/analyzer"
	"github.com/turnon/funcall/graph"
)

func main() {
	result, err := analyzer.Analyze("github.com/turnon/bookmark")
	if err != nil {
		panic(err) // internal error in pointer analysis
	}

	gd := graph.NewGraphData(result)
	gd.Process()
	gd.WriteFile("./finalHtml.html")
}
