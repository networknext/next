package storage

type MultipathVetoHandler interface {
	GetMapCopy(companyCode string) map[uint64]bool
	MultipathVetoUser(companyCode string, userHash uint64) error
	Sync() error
}
