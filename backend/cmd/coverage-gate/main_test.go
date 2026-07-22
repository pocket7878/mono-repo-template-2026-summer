package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScopedCoverageExcludesGlueGeneratedAndLintCode(t *testing.T) {
	profile := writeCoverageProfile(t, `mode: atomic
github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/usecase/todo/usecase.go:1.1,1.2 8 1
github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/usecase/todo/usecase.go:2.1,2.2 2 0
github.com/pocket7878/mono-repo-template-2026-summer/backend/cmd/server/main.go:1.1,1.2 100 0
github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/handler/api/api.gen.go:1.1,1.2 100 0
github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/lint/analyzers/analyzer.go:1.1,1.2 100 0
`)

	result, err := scopedCoverage(profile)

	require.NoError(t, err)
	require.Equal(t, coverage{covered: 8, total: 10}, result)
	require.Equal(t, 80.0, roundToTenth(float64(result.covered)*100/float64(result.total)))
}

func TestScopedCoverageRejectsMalformedAndEmptyProfiles(t *testing.T) {
	t.Run("malformed row", func(t *testing.T) {
		_, err := scopedCoverage(writeCoverageProfile(t, "mode: atomic\ninvalid\n"))
		require.ErrorContains(t, err, "invalid coverage row")
	})

	t.Run("only excluded statements", func(t *testing.T) {
		_, err := scopedCoverage(writeCoverageProfile(t, "mode: atomic\ncmd/server/main.go:1.1,1.2 1 1\n"))
		require.ErrorContains(t, err, "no statements")
	})
}

func TestRoundToTenth(t *testing.T) {
	require.Equal(t, 85.0, roundToTenth(84.96))
	require.Equal(t, 84.9, roundToTenth(84.94))
}

func writeCoverageProfile(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "coverage.out")
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))
	return path
}
