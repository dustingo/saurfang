package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNamespace
func TestNamespace(t *testing.T) {
	t.Run("AddNamespace", func(t *testing.T) {
		res := AddNamespace("10001", "game_job")
		assert.Equal(t, "game_job/10001", res)

	})
	t.Run("RemoveNamespace", func(t *testing.T) {
		res := RemoveNamespace("game_job/10001", "game_job")
		assert.Equal(t, "10001", res)
	})
}
