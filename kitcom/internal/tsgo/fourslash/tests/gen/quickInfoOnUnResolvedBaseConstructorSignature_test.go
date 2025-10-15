package fourslash_test

import (
	"testing"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/testutil"
)

func TestQuickInfoOnUnResolvedBaseConstructorSignature(t *testing.T) {
	t.Parallel()

	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `class baseClassWithConstructorParameterSpecifyingType {
    constructor(loading?: boolean) {
    }
}
class genericBaseClassInheritingConstructorFromBase<TValue> extends baseClassWithConstructorParameterSpecifyingType {
}
class classInheritingSpecializedClass extends genericBaseClassInheritingConstructorFromBase<string> {
}
new class/*1*/InheritingSpecializedClass();`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.GoToMarker(t, "1")
	f.VerifyQuickInfoExists(t)
}
