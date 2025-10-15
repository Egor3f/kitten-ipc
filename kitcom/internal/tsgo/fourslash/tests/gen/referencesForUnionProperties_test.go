package fourslash_test

import (
	"testing"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/testutil"
)

func TestReferencesForUnionProperties(t *testing.T) {
	t.Parallel()

	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `interface One {
    common: { /*one*/a: number; };
}

interface Base {
    /*base*/a: string;
    b: string;
}

interface HasAOrB extends Base {
    a: string;
    b: string;
}

interface Two {
    common: HasAOrB;
}

var x : One | Two;

x.common./*x*/a;`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.VerifyBaselineFindAllReferences(t, "one", "base", "x")
}
