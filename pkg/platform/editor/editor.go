package editor

import (
	"os"
	"path/filepath"
)

type Editorer interface {
	Info() *Editor
}

type Editor struct {
	Name string
	Path string
}

func (e *Editor) Get() *Editor {
	editor := os.Getenv("EDITOR")
	name := filepath.Base(editor)

	return &Editor{
		Name: name,
		Path: editor,
	}
}
