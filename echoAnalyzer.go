package echoAnalyzer

import (
	"net/http"

	"github.com/gostaticanalysis/analysisutil"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"
)

const doc = "echoAnalyzer parses codes written by Echo (web framework) and reports endpoint informations."

// Analyzer is ...
var Analyzer = &analysis.Analyzer{
	Name: "echoAnalyzer",
	Doc:  doc,
	Run:  run,
	Requires: []*analysis.Analyzer{
		buildssa.Analyzer,
	},
}

func run(pass *analysis.Pass) (any, error) {
	findEndpointAdditions(pass)

	return nil, nil
}

func findEndpointAdditions(pass *analysis.Pass) []*endpointAddition {
	// Define values to return
	endpoints := []*endpointAddition{}

	// Get all funcs from buildssa analyzer
	srcFuncs := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA).SrcFuncs

	// Find endpoint settings such as the `echo.GET` method
	analysisutil.InspectFuncs(srcFuncs, func(i int, instr ssa.Instruction) bool {
		// Skip if not function/method calling
		call, isCall := instr.(*ssa.Call)
		if !isCall {
			return true
		}

		// Skip if function (not method)
		if !call.Common().IsInvoke() {
			return true
		}

		receiver := call.Common().Args[0]

		switch receiver.Type() {
		case analysisutil.TypeOf(pass, "echo", "Echo"):
		case analysisutil.TypeOf(pass, "echo", "Group"):
		default:
			return true // skip if not echo receiver
		}

		var httpmethod string
		switch call.Common().Method.Name() {
		case "CONNECT":
			httpmethod = http.MethodConnect
		case "DELETE":
			httpmethod = http.MethodDelete
		case "GET":
			httpmethod = http.MethodGet
		case "HEAD":
			httpmethod = http.MethodHead
		case "OPTIONS":
			httpmethod = http.MethodOptions
		case "PATCH":
			httpmethod = http.MethodPatch
		case "POST":
			httpmethod = http.MethodPost
		case "PUT":
			httpmethod = http.MethodPut
		case "TRACE":
			httpmethod = http.MethodTrace
		default:
			return true // skip if not endpoint setting method
		}

		//pathArg := call.Common().Args[1]

		endpoints = append(
			endpoints,
			&endpointAddition{
				method:     httpmethod,
				path:       "",
				handler:    &ssa.Function{},
				middleware: []*ssa.Function{},
			},
		)

		return true
	})

	return endpoints
}

type endpointAddition struct {
	method     string
	path       string
	handler    *ssa.Function
	middleware []*ssa.Function
}
