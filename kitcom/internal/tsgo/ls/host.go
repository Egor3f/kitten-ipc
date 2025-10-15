package ls

import "efprojects.com/kitten-ipc/kitcom/internal/tsgo/sourcemap"

type Host interface {
	UseCaseSensitiveFileNames() bool
	ReadFile(path string) (contents string, ok bool)
	Converters() *Converters
	GetECMALineInfo(fileName string) *sourcemap.ECMALineInfo
}
