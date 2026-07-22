package analyzers

import (
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/lint/analyzers/repositoryentityreconstruct"
	"golang.org/x/tools/go/analysis"
)

// All returns all project-local backend analyzers.
func All() []*analysis.Analyzer {
	return []*analysis.Analyzer{repositoryentityreconstruct.Analyzer}
}
