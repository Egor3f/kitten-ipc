package fourslash_test

import (
	"testing"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash"
	. "efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash/tests/util"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/testutil"
)

func TestDocumentHighlights01(t *testing.T) {
	t.Parallel()

	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `// @Filename: a.ts
function [|f|](x: typeof [|f|]) {
    [|f|]([|f|]);
}`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.VerifyBaselineDocumentHighlights(t, nil /*preferences*/, ToAny(f.Ranges())...)
}
