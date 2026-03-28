package golang

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSerialize(t *testing.T) {
	ipc := &ipcCommon{}

	t.Run("primitives pass through", func(t *testing.T) {
		assert.Equal(t, 42, ipc.serialize(42))
		assert.Equal(t, "hello", ipc.serialize("hello"))
		assert.Equal(t, true, ipc.serialize(true))
		assert.Equal(t, 3.14, ipc.serialize(3.14))
	})

	t.Run("byte slice serializes to blob", func(t *testing.T) {
		data := []byte{0x01, 0x02, 0x03}
		result := ipc.serialize(data)
		m, ok := result.(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, "blob", m["t"])
		assert.Equal(t, "AQID", m["d"]) // base64 of {1,2,3}
	})

	t.Run("empty byte slice serializes to blob", func(t *testing.T) {
		data := []byte{}
		result := ipc.serialize(data)
		m, ok := result.(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, "blob", m["t"])
		assert.Equal(t, "", m["d"])
	})
}

func TestConvType(t *testing.T) {
	ipc := &ipcCommon{}

	t.Run("float64 to int", func(t *testing.T) {
		result := ipc.ConvType(reflect.TypeOf(0), reflect.TypeOf(0.0), float64(42))
		assert.Equal(t, 42, result)
	})

	t.Run("float64 with fractional part stays float", func(t *testing.T) {
		result := ipc.ConvType(reflect.TypeOf(0), reflect.TypeOf(0.0), float64(42.5))
		assert.Equal(t, float64(42.5), result)
	})

	t.Run("string passes through", func(t *testing.T) {
		result := ipc.ConvType(reflect.TypeOf(""), reflect.TypeOf(""), "hello")
		assert.Equal(t, "hello", result)
	})

	t.Run("bool passes through", func(t *testing.T) {
		result := ipc.ConvType(reflect.TypeOf(true), reflect.TypeOf(true), true)
		assert.Equal(t, true, result)
	})
}
