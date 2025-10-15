package fourslash_test

import (
	"testing"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/testutil"
)

func TestGoToDefinitionPrivateName(t *testing.T) {
	t.Parallel()

	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `class A {
    [|/*pnMethodDecl*/#method|]() { }
    [|/*pnFieldDecl*/#foo|] = 3;
    get [|/*pnPropGetDecl*/#prop|]() { return ""; }
    set [|/*pnPropSetDecl*/#prop|](value: string) {  }
    constructor() {
        this.[|/*pnFieldUse*/#foo|]
        this.[|/*pnMethodUse*/#method|]
        this.[|/*pnPropUse*/#prop|]
    }
}`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.VerifyBaselineGoToDefinition(t, "pnFieldUse", "pnMethodUse", "pnPropUse")
}
