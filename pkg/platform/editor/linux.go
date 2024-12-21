package editor

type LinuxEditor struct {
	*Editor
}

func NewLinuxEditor() *LinuxEditor {
	return &LinuxEditor{
		Editor: &Editor{
			Name: "vi",
			Path: "/usr/bin/vi",
		},
	}
}

func (l *LinuxEditor) Info() *Editor {
	editor := l.Get()
	if editor.Name == "" || editor.Path == "" {
		editor.Name = l.Name
		editor.Path = l.Path
	}

	return &Editor{
		Name: editor.Name,
		Path: editor.Path,
	}
}
