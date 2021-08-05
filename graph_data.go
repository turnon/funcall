package main

import (
	"strings"

	"golang.org/x/tools/go/callgraph"
)

type node struct {
	Name     string `json:"name"`
	Category int    `json:"category"`
}

type link struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type category struct {
	Name string `json:"name"`
}

type graphData struct {
	Nodes      []node     `json:"nodes"`
	Links      []link     `json:"links"`
	Categories []category `json:"categories"`

	funcSet    map[string]struct{}
	packageSet map[string]int
}

func NewGraphData() *graphData {
	gd := new(graphData)
	gd.funcSet = make(map[string]struct{})
	gd.packageSet = make(map[string]int)
	return gd
}

func (gd *graphData) addLink(edge *callgraph.Edge) {
	callerFunc := edge.Caller.Func.String()
	gd.addNodeAndCategory(callerFunc)

	calleeFunc := edge.Callee.Func.String()
	gd.addNodeAndCategory(calleeFunc)

	gd.Links = append(gd.Links, link{callerFunc, calleeFunc})
}

func (gd *graphData) addNodeAndCategory(callerFunc string) {
	if _, funcExists := gd.funcSet[callerFunc]; funcExists {
		return
	}

	gd.funcSet[callerFunc] = struct{}{}

	pkgName := packageName(callerFunc)
	pkgId, pkgExists := gd.packageSet[pkgName]

	if !pkgExists {
		pkgId = len(gd.packageSet)
		gd.packageSet[pkgName] = pkgId
		gd.Categories = append(gd.Categories, category{pkgName})
	}

	gd.Nodes = append(gd.Nodes, node{callerFunc, pkgId})
}

func withoutFunc(funcStr string) string {
	dot := strings.LastIndex(funcStr, ".")
	if dot >= 0 {
		return funcStr[:dot]
	}
	return funcStr
}

func packageName(funcStr string) string {
	declaredPos := withoutFunc(funcStr)

	if strings.HasPrefix(declaredPos, "(") {
		dot := strings.LastIndex(declaredPos, ".")
		if strings.HasPrefix(declaredPos, "(*") {
			return declaredPos[2:dot]
		}
		return declaredPos[1:dot]
	} else {
		return declaredPos
	}
}
