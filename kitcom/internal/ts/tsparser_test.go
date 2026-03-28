package ts

import (
	"path/filepath"
	"testing"

	"github.com/egor3f/kitten-ipc/kitcom/internal/api"
	"github.com/egor3f/kitten-ipc/kitcom/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTsParser(t *testing.T) {
	parser := &TypescriptApiParser{Parser: &common.Parser{}}
	absPath, err := filepath.Abs("../../../example/ts/src/index.ts")
	require.NoError(t, err)
	parser.AddFile(absPath)

	result, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, result.Endpoints, 1)

	ep := result.Endpoints[0]
	assert.Equal(t, "TsIpcApi", ep.Name)
	require.Len(t, ep.Methods, 2)

	// Div method
	div := ep.Methods[0]
	assert.Equal(t, "Div", div.Name)
	require.Len(t, div.Params, 2)
	assert.Equal(t, api.TInt, div.Params[0].Type)
	assert.Equal(t, "a", div.Params[0].Name)
	assert.Equal(t, api.TInt, div.Params[1].Type)
	assert.Equal(t, "b", div.Params[1].Name)
	require.Len(t, div.Ret, 1)
	assert.Equal(t, api.TInt, div.Ret[0].Type)

	// XorData method
	xor := ep.Methods[1]
	assert.Equal(t, "XorData", xor.Name)
	require.Len(t, xor.Params, 2)
	assert.Equal(t, api.TBlob, xor.Params[0].Type)
	assert.Equal(t, api.TBlob, xor.Params[1].Type)
	require.Len(t, xor.Ret, 1)
	assert.Equal(t, api.TBlob, xor.Ret[0].Type)
}
