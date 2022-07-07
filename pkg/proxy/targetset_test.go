package proxy

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
	"os"
	"runtime/pprof"
	"testing"
	"time"
)

var enableProfileReport = false

func TestTargetSetWithHealthCheck(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping time intensive health check test")
		return
	}

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

	go func() {
		time.Sleep(1 * time.Second)
		writeGoroutineProfileWithCount(t, 1)
	}()

	time.Sleep(3 * time.Second)

	for _, target := range goodTargets {
		require.Equal(t, false, hc.IsBlocked(target))
	}

	for _, target := range badTargets {
		require.Equal(t, true, hc.IsBlocked(target))
	}

	go func() {
		time.Sleep(5 * time.Second)
		writeGoroutineProfileWithCount(t, 2)
	}()

	time.Sleep(10 * time.Second)

	enableBadTargets = true

	go func() {
		time.Sleep(1 * time.Second)
		writeGoroutineProfileWithCount(t, 3)
	}()

	time.Sleep(3 * time.Second)

	for _, target := range goodTargets {
		require.Equal(t, false, hc.IsBlocked(target))
	}

	for _, target := range badTargets {
		require.Equal(t, false, hc.IsBlocked(target))
	}

	time.Sleep(20 * time.Second)

	writeGoroutineProfileWithCount(t, 4)
}

func writeGoroutineProfileWithCount(t *testing.T, count int) {
	// noinspection
	if !enableProfileReport {
		return
	}
	profile := pprof.Lookup("goroutine")
	f, err := os.Create(fmt.Sprintf("goroutine%d.pprof", count))
	if err != nil {
		t.Logf("cannot create pprof file for count %d: %v", count, err)
		return
	}
	if err := profile.WriteTo(f, 2); err != nil {
		t.Logf("cannot write pprof file for ocunt %d: %v", count, err)
		return
	}
}
