package fourslash_test

import (
	"testing"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/testutil"
)

func TestGoToDefinitionImportedNames10(t *testing.T) {
	t.Parallel()

	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `// @allowjs: true
// @Filename: a.js
 class /*classDefinition*/Class {
   f;
 }
 module.exports.Class = Class;
// @Filename: b.js
const { Class } = require("./a");
 [|/*classAliasDefinition*/Class|];`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.VerifyBaselineGoToDefinition(t, "classAliasDefinition")
}
