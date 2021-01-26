package svcapp

// InstallParams is used when installing a new service
type InstallParams struct {
	Name        string
	Executable  string // path to service executable
	Args        string
	DispName    string
	Description string
}
