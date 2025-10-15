package project

import (
	"context"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/lsp/lsproto"
)

type Client interface {
	WatchFiles(ctx context.Context, id WatcherID, watchers []*lsproto.FileSystemWatcher) error
	UnwatchFiles(ctx context.Context, id WatcherID) error
	RefreshDiagnostics(ctx context.Context) error
}
