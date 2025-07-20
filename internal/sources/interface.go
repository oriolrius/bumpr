package sources

type VersionSource interface {
	Name() string
	Detect(projectPath string) bool
	GetVersion(filePath string) (string, error)
	SetVersion(filePath string, newVersion string) error
	GetDefaultFileName() string
}