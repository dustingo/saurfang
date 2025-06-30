package userhandler

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v3"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm/clause"
	"log"
	"os"
	"saurfang/internal/config"
	"saurfang/internal/models/amis"
	"saurfang/internal/models/datasource"
	"saurfang/internal/models/user"
	"saurfang/internal/service/userservice"
	"saurfang/internal/tools"
	"saurfang/internal/tools/pkg"
	"strconv"
	"strings"
	"time"
)

type UserHandler struct {
	userservice.UserService
}

func NewUserHandler(svc *userservice.UserService) *UserHandler {
	return &UserHandler{*svc}
}

// Handler_ShowUserInfoByRole 查找指定role下的用户缩略信息
func (u *UserHandler) Handler_ShowUserInfoByRole(c fiber.Ctx) error {
	roleId := c.Query("roleid")
	var users []user.UserInfo
	if err := u.DB.Debug().Table("users").Preload("Roles", "id = ?", roleId).Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Where("user_roles.role_id = ?", roleId).Scan(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    users,
	})
}

// Handler_SelectRole 用于下拉菜单选择角色
func (u *UserHandler) Handler_SelectRole(c fiber.Ctx) error {
	var roles []user.Role
	if err := u.DB.Find(&roles).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": "failed to load roles",
		})
	}
	var op amis.AmisOptions
	var ops []amis.AmisOptions
	for _, role := range roles {
		op.Label = role.Name
		op.Value = int(role.ID)
		ops = append(ops, op)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data": fiber.Map{
			"options": ops,
		},
	})
}

// Handler_SetUserRole 设置用户角色
func (u *UserHandler) Handler_SetUserRole(c fiber.Ctx) error {
	userId, _ := strconv.Atoi(c.Query("userid"))
	roleId, _ := strconv.Atoi(c.Query("roleid"))
	if err := u.DB.Table("user_roles").Where("user_id = ?", userId).Update("role_id", roleId).Error; err != nil {
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

// Handler_DeleteUser 删除用户以及用户角色信息
// todo 还需补充用户权限表的清理
func (u *UserHandler) Handler_DeleteUser(c fiber.Ctx) error {
	userId, _ := strconv.Atoi(c.Query("userid"))
	tx := u.DB.Begin()
	sqls := []string{
		fmt.Sprintf("DELETE FROM user_roles WHERE user_id = %d;", userId),
		fmt.Sprintf("DELETE FROM users WHERE id = %d;", userId),
	}
	for _, sql := range sqls {
		if err := tx.Exec(sql).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  1,
				"message": err.Error(),
			})
		}
	}
	tx.Commit()
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
	})
}

// Handler_SelectPermission 变更角色权限是选择权限
func (u *UserHandler) Handler_SelectPermission(c fiber.Ctx) error {
	var permissons []user.Permission
	if err := u.DB.Table("permissions").Find(&permissons).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	var ops []amis.AmisOptions
	var op amis.AmisOptions
	for _, permisson := range permissons {
		op.Label = permisson.Name
		op.Value = int(permisson.ID)
		ops = append(ops, op)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data": fiber.Map{
			"options": ops,
		},
	})
}

func (u *UserHandler) Handler_PermissionGroupSelect(c fiber.Ctx) error {
	var permissons []user.Permission
	if err := u.DB.Table("permissions").Find(&permissons).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	var ops []amis.AmisOptions
	var op amis.AmisOptions
	for _, permisson := range permissons {
		op.Label = permisson.Group
		op.Value = int(permisson.ID)
		ops = append(ops, op)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data": fiber.Map{
			"options": ops,
		},
	})
}

// Handler_SetRolePermission 设置角色的权限|你也可以修改为设置用户的权限
func (u *UserHandler) Handler_SetRolePermission(c fiber.Ctx) error {
	roleId, _ := strconv.Atoi(c.Query("roleid"))
	permissionIds := strings.Split(c.Query("permissionids"), ",")
	var relations []user.RolePermissionRelation
	for _, pid := range permissionIds {
		id, _ := strconv.Atoi(pid)
		relations = append(relations, user.RolePermissionRelation{
			RoleID:       uint(roleId),
			PermissionID: uint(id),
		})
	}
	tx := u.DB.Begin()
	if err := tx.Table("role_permissions").
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "role_id"}, {Name: "permission_id"}},
			DoNothing: true,
		}).Create(&relations).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	tx.Commit()
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
	})
}

// Handler_UserRegister 用户注册
func (u *UserHandler) Handler_UserRegister(c fiber.Ctx) error {
	var payload user.User
	if err := c.Bind().Body(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	if payload.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": "invite code is required",
		})
	}
	if u.DB.Where("code = ? and used = 0", payload.Code).First(&user.InviteCodes{}).Error == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": "invite code is already used",
		})
	}
	apiToken, err := pkg.EncryptAES([]byte(os.Getenv("JWT_TOKEN_SECRET")), payload.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	payload.Token = apiToken
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(payload.Password), 10)
	payload.Password = string(hashedPassword)
	if err := u.DB.Create(&payload); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error,
		})
	}
	if err := u.DB.Model(&user.InviteCodes{}).Where("code = ?", payload.Code).Update("used", 1).Error; err != nil {
		log.Println("update user invite code error:", err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
	})
}

// Handler_UserLogin 用户登录
func (u *UserHandler) Handler_UserLogin(c fiber.Ctx) error {
	var userInfo user.User
	var loginStatus datasource.LoginRecords
	var payload user.LoginPayload
	if err := c.Bind().Body(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	if err := u.DB.Where("username = ?", payload.Username).First(&userInfo).Error; err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  1,
			"message": "user is not exist",
		})
	}
	if err := bcrypt.CompareHashAndPassword([]byte(userInfo.Password), []byte(payload.Password)); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  1,
			"message": "password is wrong",
		})
	}
	expireTime, _ := strconv.Atoi(os.Getenv("JWT_TOKEN_EXP"))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       userInfo.ID,
		"username": userInfo.Username,
		"role":     tools.RoleOfUser(userInfo.ID),
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(time.Duration(expireTime) * time.Second).Unix(),
	})
	tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_TOKEN_SECRET")))
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    tokenString,
		Path:     "/",
		Expires:  time.Now().Add(time.Duration(expireTime) * time.Second),
		HTTPOnly: true,
		Secure:   false,
	})
	go func(username, clientip string, ts time.Time) {
		loginStatus.Username = username
		loginStatus.ClientIP = clientip
		loginStatus.LastLogin = &ts
		config.DB.Save(&loginStatus)
	}(userInfo.Username, c.IP(), time.Now())
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "login success",
	})

}

// Handler_UserLogout 用户退出
func (u *UserHandler) Handler_UserLogout(c fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   0,
		HTTPOnly: true,
		Secure:   false,
	})
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "logout success",
	})
}
func (u *UserHandler) Handler_LoginStatus(c fiber.Ctx) error {
	cookie := c.Cookies("token")
	token, err := jwt.Parse(cookie, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_TOKEN_SECRET")), nil
	})
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  1,
			"message": "invalid token",
		})
	}

	_, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  1,
			"message": "invalid token",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "logged",
	})
}
