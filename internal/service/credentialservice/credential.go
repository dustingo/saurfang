package credentialservice

import (
	"gorm.io/gorm"
	"saurfang/internal/models/credential"
	"saurfang/internal/repository/base"
)

type CredentialService struct {
	base.BaseGormRepository[credential.UserCredential]
}

// NewHostsRepository
func NewCredentialService(db *gorm.DB) *CredentialService {
	return &CredentialService{
		BaseGormRepository: base.BaseGormRepository[credential.UserCredential]{DB: db},
	}
}
