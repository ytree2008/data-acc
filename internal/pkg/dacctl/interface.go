package dacctl

type CliContext interface {
	String(name string) string
	Int(name string) int
	Bool(name string) bool
}

type DacctlActions interface {
	CreatePersistentBuffer(c CliContext) error
	DeleteBuffer(c CliContext) error
	CreatePerJobBuffer(c CliContext) error
	ShowInstances() (string, error)
	ShowSessions() (string, error)
	ListPools() (string, error)
	ShowConfigurations() (string, error)
	ValidateJob(c CliContext) error
	RealSize(c CliContext) (string, error)
	DataIn(c CliContext) error
	Paths(c CliContext) error
	PreRun(c CliContext) error
	PostRun(c CliContext) error
	DataOut(c CliContext) error
	GenerateAnsible(c CliContext) (string, error)
}
