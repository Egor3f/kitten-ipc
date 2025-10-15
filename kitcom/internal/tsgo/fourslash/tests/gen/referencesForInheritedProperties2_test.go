package fourslash_test

import (
	"testing"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/testutil"
)

func TestReferencesForInheritedProperties2(t *testing.T) {
	t.Parallel()

	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `interface interface1 {
    /*1*/doStuff(): void;
}

interface interface2 {
    doStuff(): void;
}

interface interface2 extends interface1 {
}

class class1 implements interface2 {
    doStuff() {

    }
}

class class2 extends class1 {

}

var v: class2;
v.doStuff();`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.VerifyBaselineFindAllReferences(t, "1")
}
