package fourslash_test

import (
	"testing"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/testutil"
)

func TestGoToDefinitionFunctionOverloads(t *testing.T) {
	t.Parallel()

	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `function [|/*functionOverload1*/functionOverload|](value: number);
function /*functionOverload2*/functionOverload(value: string);
function /*functionOverloadDefinition*/functionOverload() {}

[|/*functionOverloadReference1*/functionOverload|](123);
[|/*functionOverloadReference2*/functionOverload|]("123");
[|/*brokenOverload*/functionOverload|]({});`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.VerifyBaselineGoToDefinition(t, "functionOverloadReference1", "functionOverloadReference2", "brokenOverload", "functionOverload1")
}
