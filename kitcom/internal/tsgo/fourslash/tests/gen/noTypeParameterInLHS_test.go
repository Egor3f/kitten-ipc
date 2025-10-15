package fourslash_test

import (
	"testing"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/testutil"
)

func TestNoTypeParameterInLHS(t *testing.T) {
	t.Parallel()

	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `interface I<T> { }
class C<T> {}
var /*1*/i: I<any>;
var /*2*/c: C<I>;`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.VerifyQuickInfoAt(t, "1", "var i: I<any>", "")
	f.VerifyQuickInfoAt(t, "2", "var c: C<any>", "")
}
