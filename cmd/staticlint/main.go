package main

import (
	"strings"

	"github.com/gostaticanalysis/sqlrows/passes/sqlrows"
	"github.com/reillywatson/lintservemux"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"honnef.co/go/tools/staticcheck"

	"github.com/ivanmyagkov/shortener.git/cmd/staticlint/myanalyzer"
)

func main() {
	// passesChecks contains analyzers from "golang.org/x/tools/go/analysis/passes"
	passesChecks := []*analysis.Analyzer{
		nilfunc.Analyzer,
		nilness.Analyzer,
		sigchanyzer.Analyzer,
		sortslice.Analyzer,
		stringintconv.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unusedresult.Analyzer,
		unusedwrite.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		assign.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
	}

	staticChecks := make([]*analysis.Analyzer, 100)
	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") || strings.HasPrefix(v.Analyzer.Name, "ST") {
			staticChecks = append(staticChecks, v.Analyzer)
		}
	}
	// publicChecks contains public's analyzers
	publicChecks := []*analysis.Analyzer{
		sqlrows.Analyzer,
		lintservemux.Analyzer,
	}
	//	myChecks contains custom analyzers
	myChecks := []*analysis.Analyzer{
		myanalyzer.OsExitExistsInMain,
	}
	// running analyzers
	checks := make([]*analysis.Analyzer, 100)
	checks = append(checks, passesChecks...)
	checks = append(checks, staticChecks...)
	checks = append(checks, publicChecks...)
	checks = append(checks, myChecks...)
	multichecker.Main(
		checks...,
	)

}
