package fourslash_test

import (
	"testing"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash"
	. "efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash/tests/util"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/ls"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/lsp/lsproto"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/testutil"
)

func TestCompletionsGenericIndexedAccess5(t *testing.T) {
	t.Parallel()

	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `interface CustomElements {
  'component-one': {
      foo?: string;
  },
  'component-two': {
      bar?: string;
  }
}

interface Options<T extends keyof CustomElements> {
    props?: {} & { x: CustomElements[(T extends string ? T : never) & string][] }['x'];
}

declare function f<T extends keyof CustomElements>(k: T, options: Options<T>): void;

f("component-one", {
    props: [{
        /**/
    }]
})`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.VerifyCompletions(t, "", &fourslash.CompletionsExpectedList{
		IsIncomplete: false,
		ItemDefaults: &fourslash.CompletionsExpectedItemDefaults{
			CommitCharacters: &DefaultCommitCharacters,
			EditRange:        Ignored,
		},
		Items: &fourslash.CompletionsExpectedItems{
			Exact: []fourslash.CompletionsExpectedItem{
				&lsproto.CompletionItem{
					Label:      "foo?",
					InsertText: PtrTo("foo"),
					FilterText: PtrTo("foo"),
					SortText:   PtrTo(string(ls.SortTextOptionalMember)),
				},
			},
		},
	})
}
