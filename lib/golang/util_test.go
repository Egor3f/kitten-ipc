package golang

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeErr(t *testing.T) {
	t.Run("all nil returns nil", func(t *testing.T) {
		assert.NoError(t, mergeErr(nil, nil, nil))
	})

	t.Run("single error returns it", func(t *testing.T) {
		err := fmt.Errorf("one")
		assert.EqualError(t, mergeErr(nil, err, nil), "one")
	})

	t.Run("multiple errors merged", func(t *testing.T) {
		err1 := fmt.Errorf("one")
		err2 := fmt.Errorf("two")
		result := mergeErr(err1, err2)
		assert.ErrorContains(t, result, "one")
		assert.ErrorContains(t, result, "two")
	})
}

func TestMapTypeNames(t *testing.T) {
	type Foo struct{}
	type Bar struct{}

	t.Run("maps pointer types by name", func(t *testing.T) {
		foo := &Foo{}
		bar := &Bar{}
		result := mapTypeNames([]any{foo, bar})
		assert.Equal(t, foo, result["Foo"])
		assert.Equal(t, bar, result["Bar"])
	})

	t.Run("panics on non-pointer", func(t *testing.T) {
		assert.Panics(t, func() {
			mapTypeNames([]any{Foo{}})
		})
	})
}
