package fourslash_test

import (
	"testing"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/testutil"
)

func TestGoToDefinitionSourceUnit(t *testing.T) {
	t.Parallel()

	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `// @Filename: a.ts
 //MyFile Comments
 //more comments
 /// <reference path="so/*unknownFile*/mePath.ts" />
 /// <reference path="[|b/*knownFile*/.ts|]" />

 class clsInOverload {
     static fnOverload();
     static fnOverload(foo: string);
     static fnOverload(foo: any) { }
 }

// @Filename: b.ts
/*fileB*/`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.VerifyBaselineGoToDefinition(t, "unknownFile", "knownFile")
}
