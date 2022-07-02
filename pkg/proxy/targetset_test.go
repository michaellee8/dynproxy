package proxy

import (
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
	"testing"
	"time"
)

func TestTargetSetWithHealthCheck(t *testing.T) {
	enableBadTargets := false
	goodTargets := []string{"a", "b", "c", "d"}
	badTargets := []string{"e", "f"}
	healthchecker := func(target string) bool {
		if slices.Contains(goodTargets, target) {
			return true
		}
		if slices.Contains(badTargets, target) {
			return enableBadTargets
		}
		return false
	}

	hc := NewTargetSetWithHealthCheckCustom(1*time.Second, 1, healthchecker)

	require.NoError(t, hc.Start())

	for _, target := range goodTargets {
		require.Equal(t, true, hc.Add(target))
	}

	for _, target := range badTargets {
		require.Equal(t, true, hc.Add(target))
	}

	defer func() {
		require.NoError(t, hc.Close())
	}()

	time.Sleep(3 * time.Second)

	for _, target := range goodTargets {
		require.Equal(t, false, hc.IsBlocked(target))
	}

	for _, target := range badTargets {
		require.Equal(t, true, hc.IsBlocked(target))
	}

	enableBadTargets = true

	time.Sleep(3 * time.Second)

	for _, target := range goodTargets {
		require.Equal(t, false, hc.IsBlocked(target))
	}

	for _, target := range badTargets {
		require.Equal(t, false, hc.IsBlocked(target))
	}

}
