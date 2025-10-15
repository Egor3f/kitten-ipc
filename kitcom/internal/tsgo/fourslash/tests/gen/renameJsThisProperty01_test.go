package fourslash_test

import (
	"testing"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/testutil"
)

func TestRenameJsThisProperty01(t *testing.T) {
	t.Parallel()

	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `// @allowJs: true
// @Filename: a.js
function bar() {
    [|this.[|{| "contextRangeIndex": 0 |}x|] = 10;|]
}
var t = new bar();
[|t.[|{| "contextRangeIndex": 2 |}x|] = 11;|]`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.VerifyBaselineRenameAtRangesWithText(t, nil /*preferences*/, "x")
}
