package vfsmock

import "efprojects.com/kitten-ipc/kitcom/internal/tsgo/vfs"

// Wrap wraps a vfs.FS and returns a FSMock which calls it.
func Wrap(fs vfs.FS) *FSMock {
	return &FSMock{
		DirectoryExistsFunc:           fs.DirectoryExists,
		FileExistsFunc:                fs.FileExists,
		GetAccessibleEntriesFunc:      fs.GetAccessibleEntries,
		ReadFileFunc:                  fs.ReadFile,
		RealpathFunc:                  fs.Realpath,
		RemoveFunc:                    fs.Remove,
		ChtimesFunc:                   fs.Chtimes,
		StatFunc:                      fs.Stat,
		UseCaseSensitiveFileNamesFunc: fs.UseCaseSensitiveFileNames,
		WalkDirFunc:                   fs.WalkDir,
		WriteFileFunc:                 fs.WriteFile,
	}
}
