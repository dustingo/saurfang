package credentialhandler

import (
	"saurfang/internal/models/credential"
	"saurfang/internal/repository/base"
	"saurfang/internal/tools/pkg"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm/clause"
)

type CredentialHandler struct {
	base.BaseGormRepository[credential.UserCredential]
}

// Handler_CreateUserCredential 创建你用户ak、sk
// 每个userid都只能有一条记录
func (d *CredentialHandler) Handler_CreateUserCredential(c fiber.Ctx) error {
	userid, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "request error", err.Error(), fiber.Map{})
	}
	if userid <= 0 {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "request error", "userid should be positive", fiber.Map{})
	}
	credtials, err := pkg.GenerateAKSKPair(uint(userid))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to generate ak/sk pair", err.Error(), fiber.Map{})
	}
	result := d.DB.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&credtials)
	if result.Error != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to create user credential", result.Error.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", nil)
}

func (d *CredentialHandler) Handler_DeleteUserCredential(c fiber.Ctx) error {
	userid, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "request error", err.Error(), fiber.Map{})
	}
	if userid <= 0 {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "request error", "userid should be positive", fiber.Map{})
	}
	if err := d.DB.Table("user_credentials").Where("user_id = ?", uint(userid)).Delete(&credential.UserCredential{}).Error; err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to delete user credential", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", nil)
}

func (d *CredentialHandler) Handler_ShowUserCredential(c fiber.Ctx) error {
	credentials, err := d.List()
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to show user credential", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", credentials)
}
func (d *CredentialHandler) Handler_SetUserCredentialStatus(c fiber.Ctx) error {
	userid, err := strconv.Atoi(c.Query("userid"))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "request error", err.Error(), fiber.Map{})
	}
	if userid <= 0 {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "request error", "userid should be positive", fiber.Map{})
	}
	status := c.Query("status")
	switch status {
	case "active":
		if err := d.DB.Table("user_credentials").Where("user_id = ? ", uint(userid)).Update("status", "active").Error; err != nil {
			return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to set user credential status", err.Error(), fiber.Map{})
		}
	case "inactive":
		if err := d.DB.Table("user_credentials").Where("user_id = ? ", uint(userid)).Update("status", "inactive").Error; err != nil {
			return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to set user credential status", err.Error(), fiber.Map{})
		}
	default:
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "request error", "invalid status", fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", nil)
}
