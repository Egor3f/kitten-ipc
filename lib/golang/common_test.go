package golang

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testEndpoint struct{}

func (e *testEndpoint) Hello(name string) (string, error) {
	return "hello " + name, nil
}

func TestFindMethod(t *testing.T) {
	ipc := &ipcCommon{
		localApis: mapTypeNames([]any{&testEndpoint{}}),
	}

	t.Run("valid method", func(t *testing.T) {
		method, err := ipc.findMethod("testEndpoint.Hello")
		assert.NoError(t, err)
		assert.True(t, method.IsValid())
	})

	t.Run("invalid format - no dot", func(t *testing.T) {
		_, err := ipc.findMethod("Hello")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid method")
	})

	t.Run("unknown endpoint", func(t *testing.T) {
		_, err := ipc.findMethod("Unknown.Hello")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "endpoint not found")
	})

	t.Run("unknown method", func(t *testing.T) {
		_, err := ipc.findMethod("testEndpoint.Unknown")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "method not found")
	})
}
