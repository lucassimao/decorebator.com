package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAnalyticsInvalidationScansEveryVersionedPageVariant(t *testing.T) {
	t.Parallel()
	patterns := analyticsCacheInvalidationPatterns(11, 22)

	require.Contains(t, patterns, "analytics:mastery:*:11:22:*")
	require.Contains(t, patterns, "analytics:progress-summary:v*:11:*")
	require.Contains(t, patterns, "analytics:progress:11:22:*")
	require.Contains(t, patterns, "analytics:practice-time:11:22:*")
	require.Contains(t, patterns, "analytics:historical-box-distribution:11:22:*")
}
