package editor

type WindowsEditor struct {
	*Editor
}

func NewWindowsEditor() *WindowsEditor {
	return &WindowsEditor{
		Editor: &Editor{
			Name: "notepad",
			Path: "C:\\Windows\\System32\\notepad.exe",
		},
	}
}

func (w *WindowsEditor) Info() *Editor {
	editor := w.Get()
	if editor.Name == "" || editor.Path == "" {
		editor.Name = w.Name
		editor.Path = w.Path
	}

	return &Editor{
		Name: editor.Name,
		Path: editor.Path,
	}
}
