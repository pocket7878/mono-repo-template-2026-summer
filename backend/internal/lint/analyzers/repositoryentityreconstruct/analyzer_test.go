package repositoryentityreconstruct_test

import (
	"testing"

	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/lint/analyzers/repositoryentityreconstruct"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testData := analysistest.TestData()
	analysistest.Run(t, testData, repositoryentityreconstruct.Analyzer,
		"a/internal/repository",
		"a/internal/usecase",
	)
}
