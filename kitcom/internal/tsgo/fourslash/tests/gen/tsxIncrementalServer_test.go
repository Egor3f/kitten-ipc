package fourslash_test

import (
	"testing"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/testutil"
)

func TestTsxIncrementalServer(t *testing.T) {
	t.Parallel()

	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `/**/`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.GoToMarker(t, "")
	f.Insert(t, "<")
	f.Insert(t, "div")
	f.Insert(t, " ")
	f.Insert(t, " id")
	f.Insert(t, "=")
	f.Insert(t, "\"foo")
	f.Insert(t, "\"")
	f.Insert(t, ">")
}
