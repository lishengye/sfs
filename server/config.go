package server

type Config struct {
	Port      int16
	Directory string
}

func NewConfig(port int16, directory string) Config {
	return Config{
		Port:      port,
		Directory: directory,
	}
}
