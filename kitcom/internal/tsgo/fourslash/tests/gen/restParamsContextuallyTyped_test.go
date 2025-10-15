package fourslash_test

import (
	"testing"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/testutil"
)

func TestRestParamsContextuallyTyped(t *testing.T) {
	t.Parallel()

	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `var foo: Function = function (/*1*/a, /*2*/b, /*3*/c) { };`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.VerifyQuickInfoAt(t, "1", "(parameter) a: any", "")
	f.VerifyQuickInfoAt(t, "2", "(parameter) b: any", "")
	f.VerifyQuickInfoAt(t, "3", "(parameter) c: any", "")
}
