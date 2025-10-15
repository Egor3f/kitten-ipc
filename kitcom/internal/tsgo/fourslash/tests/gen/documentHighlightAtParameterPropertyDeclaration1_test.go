package fourslash_test

import (
	"testing"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash"
	. "efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash/tests/util"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/testutil"
)

func TestDocumentHighlightAtParameterPropertyDeclaration1(t *testing.T) {
	t.Parallel()

	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `// @Filename: file1.ts
class Foo {
    constructor(private [|privateParam|]: number,
        public [|publicParam|]: string,
        protected [|protectedParam|]: boolean) {

        let localPrivate = [|privateParam|];
        this.[|privateParam|] += 10;

        let localPublic = [|publicParam|];
        this.[|publicParam|] += " Hello!";

        let localProtected = [|protectedParam|];
        this.[|protectedParam|] = false;
    }
}`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.VerifyBaselineDocumentHighlights(t, nil /*preferences*/, ToAny(f.Ranges())...)
}
