package editor

type DarwinEditor struct {
	*Editor
}

func NewDarwinEditor() *DarwinEditor {
	return &DarwinEditor{
		Editor: &Editor{
			Name: "open",
			Path: "/usr/bin/open",
		},
	}
}

func (d *DarwinEditor) Info() *Editor {
	editor := d.Get()
	if editor.Name == "" || editor.Path == "" {
		editor.Name = d.Name
		editor.Path = d.Path
	}

	return &Editor{
		Name: editor.Name,
		Path: editor.Path,
	}
}
