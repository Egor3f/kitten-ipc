package sourcemap

import "efprojects.com/kitten-ipc/kitcom/internal/tsgo/core"

type Source interface {
	Text() string
	FileName() string
	ECMALineMap() []core.TextPos
}
