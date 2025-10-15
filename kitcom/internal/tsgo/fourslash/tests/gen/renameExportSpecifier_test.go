package fourslash_test

import (
	"testing"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/core"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/fourslash"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/ls"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/testutil"
)

func TestRenameExportSpecifier(t *testing.T) {
	t.Parallel()

	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `// @Filename: a.ts
const name = {};
export { name as name/**/ };
// @Filename: b.ts
import { name } from './a';
const x = name.toString();`
	f := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	f.VerifyBaselineRename(t, &ls.UserPreferences{UseAliasesForRename: core.TSFalse}, "")
}
