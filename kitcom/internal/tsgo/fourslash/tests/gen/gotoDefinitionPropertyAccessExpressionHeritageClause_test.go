package fourslash_test

import (
	"testing"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/testutil"
)

func TestGotoDefinitionPropertyAccessExpressionHeritageClause(t *testing.T) {
	t.Parallel()

	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `class B {}
function foo() {
    return {/*refB*/B: B};
}
class C extends (foo()).[|/*B*/B|] {}
class C1 extends foo().[|/*B1*/B|] {}`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.VerifyBaselineGoToDefinition(t, "B", "B1")
}
