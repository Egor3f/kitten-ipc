package fourslash_test

import (
	"testing"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/testutil"
)

func TestQuickInfoModuleVariables(t *testing.T) {
	t.Parallel()
	t.Skip()
	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `var x = 1;
module M {
    export var x = 2;
    console.log(/*1*/x); // 2
}
module M {
    console.log(/*2*/x); // 2
}
module M {
    var x = 3;
    console.log(/*3*/x); // 3
}`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.VerifyQuickInfoAt(t, "1", "var M.x: number", "")
	f.VerifyQuickInfoAt(t, "2", "var M.x: number", "")
	f.VerifyQuickInfoAt(t, "3", "var x: number", "")
}
