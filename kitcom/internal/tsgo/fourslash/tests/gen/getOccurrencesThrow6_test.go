package fourslash_test

import (
	"testing"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash"
	. "efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash/tests/util"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/testutil"
)

func TestGetOccurrencesThrow6(t *testing.T) {
	t.Parallel()

	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `[|throw|] 100;

try {
    throw 0;
    var x = () => { throw 0; };
}
catch (y) {
    var x = () => { throw 0; };
    [|throw|] 200;
}
finally {
    [|throw|] 300;
}`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.VerifyBaselineDocumentHighlights(t, nil /*preferences*/, ToAny(f.Ranges())...)
}
