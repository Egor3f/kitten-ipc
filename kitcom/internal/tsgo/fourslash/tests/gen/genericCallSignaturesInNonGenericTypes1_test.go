package fourslash_test

import (
	"testing"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/testutil"
)

func TestGenericCallSignaturesInNonGenericTypes1(t *testing.T) {
	t.Parallel()

	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `interface WrappedObject<T> { }
interface WrappedArray<T> { }
interface Underscore {
    <T>(list: T[]): WrappedArray<T>;
    <T>(obj: T): WrappedObject<T>;
}
var _: Underscore;
var a: number[];
var /**/b = _(a); `
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.VerifyQuickInfoAt(t, "", "var b: WrappedArray<number>", "")
}
