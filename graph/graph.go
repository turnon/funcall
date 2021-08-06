package graph

import (
	_ "embed"

	"encoding/json"
	"io/ioutil"
	"regexp"
	"strings"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/pointer"
)

//go:embed verd.html
var tmpl string

var toReplace = regexp.MustCompile(`(?s)//start-sub.*//end-sub`)

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

	pointerResult *pointer.Result
	funcSet       map[string]struct{}
	packageSet    map[string]int
	callSet       map[link]struct{}
}

func NewGraphData(result *pointer.Result) *graphData {
	gd := new(graphData)
	gd.pointerResult = result
	gd.funcSet = make(map[string]struct{})
	gd.packageSet = make(map[string]int)
	gd.callSet = make(map[link]struct{})
	return gd
}

func (gd *graphData) Process(packageNames []string) {
	edgeConcernedPkg := func(edge *callgraph.Edge) bool {
		edgeStr := edge.String()
		for _, packageName := range packageNames {
			if strings.Contains(edgeStr, packageName) {
				return true
			}
		}
		return false
	}

	for _, nodes := range gd.pointerResult.CallGraph.Nodes {
		for _, edge := range nodes.Out {
			if edgeConcernedPkg(edge) {
				gd.addLink(edge)
			}
		}
	}
}

func (gd *graphData) WriteFile(targetFilePath string) {
	bytes, _ := json.Marshal(gd)
	finalHtml := toReplace.ReplaceAllLiteralString(tmpl, string(bytes))
	ioutil.WriteFile(targetFilePath, []byte(finalHtml), 0644)
}

func (gd *graphData) addLink(edge *callgraph.Edge) {
	callerFunc := edge.Caller.Func.String()
	gd.addNodeAndCategory(callerFunc)

	calleeFunc := edge.Callee.Func.String()
	gd.addNodeAndCategory(calleeFunc)

	link := link{callerFunc, calleeFunc}
	if _, linkExists := gd.callSet[link]; !linkExists {
		gd.callSet[link] = struct{}{}
		gd.Links = append(gd.Links, link)
	}
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
