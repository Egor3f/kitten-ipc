package fourslash_test

import (
	"testing"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/testutil"
)

func TestGoToDefinitionMetaProperty(t *testing.T) {
	t.Parallel()

	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `// @Filename: /a.ts
im/*1*/port.met/*2*/a;
function /*functionDefinition*/f() { n/*3*/ew.[|t/*4*/arget|]; }
// @Filename: /b.ts
im/*5*/port.m;
class /*classDefinition*/c { constructor() { n/*6*/ew.[|t/*7*/arget|]; } }`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.VerifyBaselineGoToDefinition(t, "1", "2", "3", "4", "5", "6", "7")
}
