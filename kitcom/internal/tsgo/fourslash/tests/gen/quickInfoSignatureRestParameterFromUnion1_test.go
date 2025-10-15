package fourslash_test

import (
	"testing"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/testutil"
)

func TestQuickInfoSignatureRestParameterFromUnion1(t *testing.T) {
	t.Parallel()
	t.Skip()
	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `declare const rest:
  | ((v: { a: true }, ...rest: string[]) => unknown)
  | ((v: { b: true }) => unknown);

/**/rest({ a: true, b: true }, "foo", "bar");`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.VerifyQuickInfoAt(t, "", "const rest: (v: {\n    a: true;\n} & {\n    b: true;\n}, ...rest: string[]) => unknown", "")
}
