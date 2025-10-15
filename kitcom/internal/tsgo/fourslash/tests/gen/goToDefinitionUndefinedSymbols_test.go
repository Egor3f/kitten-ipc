package fourslash_test

import (
	"testing"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/testutil"
)

func TestGoToDefinitionUndefinedSymbols(t *testing.T) {
	t.Parallel()

	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `some/*undefinedValue*/Variable;
var a: some/*undefinedType*/Type;
var x = {}; x.some/*undefinedProperty*/Property;
var a: any; a.some/*unkownProperty*/Property;`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.VerifyBaselineGoToDefinition(t, f.MarkerNames()...)
}
