package main

import (
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/lint/analyzers"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(analyzers.All()...)
}
