package golang

import (
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewParent(t *testing.T) {
	t.Run("socket argument in command", func(t *testing.T) {
		cmd := exec.Command("/bin/sh", ipcSocketArg, "/tmp/kek")
		_, err := NewParent(cmd)
		assert.Error(t, err)
	})

	t.Run("nonexistent binary", func(t *testing.T) {
		cmd := exec.Command("/nonexistent/binary")
		p, err := NewParent(cmd)
		assert.NoError(t, err)
		assert.Error(t, p.Start())
	})

	t.Run("connection timeout", func(t *testing.T) {
		cmd := exec.Command("../testdata/sleep15.sh")
		p, err := NewParent(cmd)
		assert.NoError(t, err)
		assert.Error(t, p.Start())
	})

	t.Run("child finished before accepting connection", func(t *testing.T) {
		cmd := exec.Command("../testdata/sleep3.sh")
		p, err := NewParent(cmd)
		assert.NoError(t, err)
		start := time.Now()
		assert.Error(t, p.Start())
		assert.WithinDuration(t, time.Now(), start, time.Second*4)
	})
}
