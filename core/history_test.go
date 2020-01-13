package core_test

import (
	"testing"

	"github.com/networknext/backend/core"
	"github.com/stretchr/testify/assert"
)

func TestHistory(t *testing.T) {
	t.Run("HistoryMax()", func(t *testing.T) {
		t.Run("returns the max value in the array", func(t *testing.T) {
			history := []float32{1, 2, 3, 4, 5, 4, 3, 2, 1}
			assert.Equal(t, float32(5), core.HistoryMax(history))
		})
	})

	t.Run("HistoryNotSet()", func(t *testing.T) {
		t.Run("returns a []float32 the size of HistorySize containing nothing but InvalidHistoryValue", func(t *testing.T) {
			history := core.HistoryNotSet()
			assert.Len(t, history, core.HistorySize)
			for i := 0; i < core.HistorySize; i++ {
				assert.Equal(t, core.InvalidHistoryValue, int(history[i]))
			}
		})
	})

	t.Run("HistoryMean()", func(t *testing.T) {
		t.Run("returns a float32 that is the average of the input", func(t *testing.T) {
			history := []float32{1, 2, 3, 4, 5, 4, 3, 2, 1}
			assert.Equal(t, float32(25.0/9.0), core.HistoryMean(history))
		})
	})
}
