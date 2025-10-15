package fourslash_test

import (
	"testing"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/testutil"
)

func TestQuickInfoForIn(t *testing.T) {
	t.Parallel()

	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `var obj;
for (var /**/p in obj) { }`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.VerifyQuickInfoAt(t, "", "var p: string", "")
}
