package middleware

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v3"
	"log"
	"os"
	"saurfang/internal/config"
	"strings"
)

func UserAuth() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		log.Println("Request Path: ", string(ctx.Request().URI().Path()))
		if strings.HasPrefix(string(ctx.Request().URI().Path()), "/api/v1/common") {
			log.Println("white list /api/v1/common")
			return ctx.Next()
		}
		ck := ctx.Cookies("token")
		if ck == "" {
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  1,
				"message": "unauthorized",
			})
		}
		token, err := jwt.Parse(ck, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_TOKEN_SECRET")), nil
		})
		if err != nil || !token.Valid {
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  1,
				"message": "unauthorized",
			})
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  1,
				"message": "unauthorized",
			})
		}
		role := claims["role"].(interface{})
		var perm string
		pats := strings.Split(string(ctx.Request().URI().Path()), "/")
		if len(pats) >= 4 {
			perm = strings.Join(pats[:4], "/")
		} else {
			perm = string(ctx.Request().URI().Path())
		}
		log.Println("perm: ", perm)
		if hasPermission(uint(role.(float64)), perm) {
			return ctx.Next()
		} else {
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  1,
				"message": "unauthorized",
			})
		}
	}
}
func hasPermission(roleid uint, path string) bool {
	type UserPermission struct {
		ID    uint   ` json:"id"`
		Name  string `json:"name"`
		Group string `json:"group"`
	}
	var ups []UserPermission
	// 查找role_permissions里是否有请求权限
	if err := config.DB.Debug().Raw("SELECT p.name,p.group,p.`id`  from permissions p JOIN  role_permissions rp ON rp.permission_id = p.id where rp.role_id = ?", roleid).Scan(&ups).Error; err != nil {
		log.Println("query error: ", err.Error())
		return false
	}
	for _, g := range ups {
		if g.Name == path {
			return true
		}
	}
	return false
}
