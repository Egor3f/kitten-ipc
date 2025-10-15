package fourslash_test

import (
	"testing"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/testutil"
)

func TestFindAllRefsParameterPropertyDeclaration1(t *testing.T) {
	t.Parallel()

	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `class Foo {
    constructor(private /*1*/privateParam: number) {
        let localPrivate = privateParam;
        this.privateParam += 10;
    }
}`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.VerifyBaselineFindAllReferences(t, "1")
}
