package fourslash_test

import (
	"testing"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/testutil"
)

func TestGetJavaScriptCompletions22(t *testing.T) {
	t.Parallel()

	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `// @allowNonTsExtensions: true
// @Filename: file.js
const abc = {};
({./*1*/});`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.GoToMarker(t, "1")
	f.Insert(t, ".")
	f.VerifyCompletions(t, nil, nil)
}
