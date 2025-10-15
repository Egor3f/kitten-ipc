package fourslash_test

import (
	"testing"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/testutil"
)

func TestJsDocPropertyDescription7(t *testing.T) {
	t.Parallel()
	t.Skip()
	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `class StringClass {
    /** Something generic */
    static [p: string]: any;
}
function stringClass(e: typeof StringClass) {
    console.log(e./*stringClass*/anything);
}`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.VerifyQuickInfoAt(t, "stringClass", "(index) StringClass[string]: any", "Something generic")
}
