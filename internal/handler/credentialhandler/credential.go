package credentialhandler

import (
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm/clause"
	"saurfang/internal/models/credential"
	"saurfang/internal/repository/base"
	"saurfang/internal/tools/pkg"
	"strconv"
)

type CredentialHandler struct {
	base.BaseGormRepository[credential.UserCredential]
	//credentialservice.CredentialService
}

//func NewCredentialHandler(svc *credentialservice.CredentialService) *CredentialHandler {
//	return &CredentialHandler{*svc}
//}

// Handler_CreateUserCredential 创建你用户ak、sk
// 每个userid都只能有一条记录
func (d *CredentialHandler) Handler_CreateUserCredential(c fiber.Ctx) error {
	userid, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	if userid <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": "userid should be positive",
		})
	}
	credtials, err := pkg.GenerateAKSKPair(uint(userid))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	result := d.DB.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&credtials)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": result.Error.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
	})
}

func (d *CredentialHandler) Handler_DeleteUserCredential(c fiber.Ctx) error {
	userid, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	if userid <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": "userid should be positive",
		})
	}
	if err := d.DB.Table("user_credentials").Where("user_id = ?", uint(userid)).Delete(&credential.UserCredential{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
	})
}

func (d *CredentialHandler) Handler_ShowUserCredential(c fiber.Ctx) error {
	credentials, err := d.List()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    credentials,
	})
}
func (d *CredentialHandler) Handler_SetUserCredentialStatus(c fiber.Ctx) error {
	userid, err := strconv.Atoi(c.Query("userid"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	if userid <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": "userid should be positive",
		})
	}
	status := c.Query("status")
	switch status {
	case "active":
		if err := d.DB.Table("user_credentials").Where("user_id = ? ", uint(userid)).Update("status", "active").Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  1,
				"message": err.Error(),
			})
		}
	case "inactive":
		if err := d.DB.Table("user_credentials").Where("user_id = ? ", uint(userid)).Update("status", "inactive").Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  1,
				"message": err.Error(),
			})
		}
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": "invalid status",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
	})
}
