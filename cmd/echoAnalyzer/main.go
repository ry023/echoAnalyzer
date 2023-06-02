package main

import (
	"github.com/ry023/echoAnalyzer"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() { unitchecker.Main(echoAnalyzer.Analyzer) }
