package echoAnalyzer

import (
	"fmt"
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
	endpoints := findEndpointSettings(pass)

	for _, e := range endpoints {
		findBoundParam(pass, e)
	}

	return nil, nil
}

func findBoundParam(pass *analysis.Pass, endpoint *endpointSetting) {
	eachInstruction([]*ssa.Function{endpoint.handler}, func(inst ssa.Instruction) {
	})
}

func findEndpointSettings(pass *analysis.Pass) []*endpointSetting {
	// Define values to return
	endpoints := []*endpointSetting{}

	// Get all funcs from buildssa analyzer
	srcFuncs := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA).SrcFuncs

	// Find endpoint settings such as the `echo.GET` method
	//analysisutil.InspectFuncs(srcFuncs, func(i int, instr ssa.Instruction) bool {
	eachInstruction(srcFuncs, func(instr ssa.Instruction) {
		// Skip if not method calling
		if !isStructMethodCall(instr) {
			return
		}

		call := instr.(*ssa.Call)
		receiver := call.Common().Signature().Recv()

		// Skip
		if len(call.Common().Args) == 0 {
			return
		}


		switch receiver.Type().String() {
		case analysisutil.TypeOf(pass, "github.com/labstack/echo/v4", "*Echo").String():
		case analysisutil.TypeOf(pass, "github.com/labstack/echo/v4", "*Group").String():
		default:
			return // skip if not echo receiver
		}

		callee := call.Common().StaticCallee()
		if callee == nil {
			return
		}

		var httpmethod string
		switch callee.Name() {
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
			return // skip if not endpoint setting method
		}

		path, ok := call.Common().Args[1].(*ssa.Const)
		if !ok {
			pass.Reportf(call.Pos(), "Path parameter must be constant (*ssa.Const) to parse.")
			return
		}

		// Parse handler
		handler, err := getFunctionFromArg(call.Common().Args[2])
		if err != nil {
			pass.Reportf(call.Pos(), "cannot parse 2nd arg: %v", err)
			return
		}

		endpoints = append(
			endpoints,
			&endpointSetting{
				method:     httpmethod,
				path:       path.Value.String(),
				handler:    handler,
				middleware: []*ssa.Function{},
			},
		)

		return
	})

	return endpoints
}

func getFunctionFromArg(arg ssa.Value) (*ssa.Function, error) {
	ct, ok := arg.(*ssa.ChangeType)
	if !ok {
		return nil, fmt.Errorf("argument value (%v) is not *ssa.ChangeType", ct)
	}
	switch v := ct.X.(type) {
	case *ssa.Function:
		return v, nil
	case *ssa.MakeClosure:
		if fn, ok := v.Fn.(*ssa.Function); ok {
			return fn, nil
		}
	}
	return nil, fmt.Errorf("argument may be not function call")
}

type endpointSetting struct {
	method     string
	path       string
	handler    *ssa.Function
	middleware []*ssa.Function
}
