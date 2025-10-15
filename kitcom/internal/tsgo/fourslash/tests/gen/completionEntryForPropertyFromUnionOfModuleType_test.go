package fourslash_test

import (
	"testing"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash"
	. "efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash/tests/util"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/lsp/lsproto"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/testutil"
)

func TestCompletionEntryForPropertyFromUnionOfModuleType(t *testing.T) {
	t.Parallel()

	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `module E {
    export var n = 1;
    export var x = 0;
}
module F {
    export var n = 1;
    export var y = 0;
}
var q: typeof E | typeof F;
var j = q./*1*/`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.VerifyCompletions(t, "1", &fourslash.CompletionsExpectedList{
		IsIncomplete: false,
		ItemDefaults: &fourslash.CompletionsExpectedItemDefaults{
			CommitCharacters: &DefaultCommitCharacters,
			EditRange:        Ignored,
		},
		Items: &fourslash.CompletionsExpectedItems{
			Exact: []fourslash.CompletionsExpectedItem{
				&lsproto.CompletionItem{
					Label:  "n",
					Detail: PtrTo("(property) n: number"),
				},
			},
		},
	})
}
