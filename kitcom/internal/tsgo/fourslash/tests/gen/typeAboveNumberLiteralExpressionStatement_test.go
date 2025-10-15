package fourslash_test

import (
	"testing"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/testutil"
)

func TestTypeAboveNumberLiteralExpressionStatement(t *testing.T) {
	t.Parallel()

	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `
// foo
1;`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.GoToBOF(t)
	f.Insert(t, "var x;\n")
}
