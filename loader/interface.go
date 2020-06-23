package loader

type Loader interface {
	Load(path string, to interface{}) error
}
