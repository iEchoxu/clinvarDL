package configs

type Configer interface {
	Write(path string) error
	Read(path string) (Configer, error)
}
