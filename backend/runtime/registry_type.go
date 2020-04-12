package runtime


type RegistryType int

//go:generate enumer -type  RegistryType -linecomment -trimprefix Registry
const (
	RegistryMaster RegistryType = iota
	RegistryEtcd
	RegistryService
	RegistryUnknown
)

