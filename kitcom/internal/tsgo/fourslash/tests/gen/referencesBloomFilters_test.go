package fourslash_test

import (
	"testing"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/testutil"
)

func TestReferencesBloomFilters(t *testing.T) {
	t.Parallel()

	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `// @Filename: declaration.ts
var container = { /*1*/searchProp : 1 };
// @Filename: expression.ts
function blah() { return (1 + 2 + container.searchProp()) === 2;  };
// @Filename: stringIndexer.ts
function blah2() { container["searchProp"] };
// @Filename: redeclaration.ts
container = { "searchProp" : 18 };`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.VerifyBaselineFindAllReferences(t, "1")
}
