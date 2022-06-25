package pickset

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestPickableSet(t *testing.T) {
	ps := NewPickableSet[string]()
	p := NewPicker(ps)
	require.Equal(t, true, ps.Add("a"))
	require.Equal(t, false, ps.Add("a"))
	require.Equal(t, 1, ps.Len())
	require.Equal(t, true, ps.Add("b"))
	require.Equal(t, true, ps.Add("c"))
	require.Equal(t, true, ps.Add("d"))
	require.Equal(t, true, ps.Add("e"))
	require.Equal(t, true, ps.Block("c"))
	require.Equal(t, false, ps.Block("c"))
	require.Equal(t, true, ps.Block("d"))
	require.Equal(t, false, ps.Block("zzz"))
	require.Equal(t, false, ps.Unblock("zzz"))
	for i := 0; i < 3; i++ {
		{
			picked, err := p.Pick()
			require.NoError(t, err)
			require.Equal(t, "a", picked)
		}
		{
			picked, err := p.Pick()
			require.NoError(t, err)
			require.Equal(t, "b", picked)
		}
		{
			picked, err := p.Pick()
			require.NoError(t, err)
			require.Equal(t, "e", picked)
		}
	}

	require.Equal(t, true, ps.Block("a"))
	require.Equal(t, true, ps.Block("b"))
	require.Equal(t, true, ps.Block("e"))
	{
		_, err := p.Pick()
		require.ErrorIs(t, err, ErrNoElementAvailableForPicking)
	}

	// Testing scheduled blocking logic
	require.Equal(t, true, ps.IsBlocked("e"))
	require.Equal(t, false, ps.BlockForDuration("e", time.Millisecond*200))
	require.Equal(t, true, ps.Unblock("e"))
	require.Equal(t, false, ps.IsBlocked("e"))
	require.Equal(t, true, ps.BlockForDuration("e", time.Millisecond*200))
	require.Equal(t, true, ps.IsBlocked("e"))
	require.Equal(t, false, ps.BlockForDuration("e", time.Millisecond*100))
	<-time.After(time.Millisecond * 300)
	require.Equal(t, false, ps.IsBlocked("e"))

	ap := NewAllPicker(ps)
	for i := 0; i < 3; i++ {
		{
			picked, err := ap.Pick()
			require.NoError(t, err)
			require.Equal(t, "a", picked)
		}
		{
			picked, err := ap.Pick()
			require.NoError(t, err)
			require.Equal(t, "b", picked)
		}
		{
			picked, err := ap.Pick()
			require.NoError(t, err)
			require.Equal(t, "c", picked)
		}
		{
			picked, err := ap.Pick()
			require.NoError(t, err)
			require.Equal(t, "d", picked)
		}
		{
			picked, err := ap.Pick()
			require.NoError(t, err)
			require.Equal(t, "e", picked)
		}
	}

}
