package settings

const (
	Storage = "output"
)

type OutputSettings struct {
	Storage string `yaml:"storage"`
}

func NewOutputSettings() *OutputSettings {
	return &OutputSettings{
		Storage: Storage,
	}
}
