package middleware

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v3"
	"net/http"
	"os"
	"saurfang/internal/config"
	"saurfang/internal/tools/pkg"
	"strconv"
	"strings"
)

func UserAuth() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		requestPath := string(ctx.Request().URI().Path())
		// 确保用户登录等通用类不受限制
		if strings.HasPrefix(requestPath, "/api/v1/common") {
			return ctx.Next()
		}
		// 校验cookie,如果存在cookie就校验cookie jwt；
		// 如果没有cookie则校验ak/sk
		ck := ctx.Cookies("token")
		if ck == "" {
			//校验ak/sk
			accessKey := ctx.Get("X-Access-Key")
			signature := ctx.Get("X-Signature")
			timestamp := ctx.Get("X-Timestamp")
			// 凭证信息都存在
			perm := formatRequestPath(requestPath)
			if accessKey != "" && signature != "" && timestamp != "" {
				userid, code, ok := pkg.VerifySignature(accessKey, signature, ctx.Method(), perm, timestamp)
				if !ok {
					// 校验失败
					return ctx.Status(code).JSON(fiber.Map{
						"status":  1,
						"message": http.StatusText(code),
					})
				}
				// 凭证校验通过,继续进行权限校验
				roleid, err := pkg.GetRoleOfUser(userid)
				if err != nil {
					return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
						"status":  1,
						"message": err.Error(),
					})
				}
				//perm := formatRequestPath(requestPath)
				if hasPermission(roleid, perm) {
					ctx.Request().Header.Set("X-Request-User", strconv.Itoa(int(userid)))
					//ctx.Set("X-Request-User", (claims["username"].(interface{}).(string)))
					return ctx.Next()
				} else {
					return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
						"status":  1,
						"message": "unauthorized",
					})
				}

			} else {
				// 凭证不存在，jwt也不存在
				return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"status":  1,
					"message": "unauthorized",
				})
			}
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
		perm := formatRequestPath(requestPath)
		if hasPermission(uint(role.(float64)), perm) {
			ctx.Request().Header.Set("X-Request-User", (claims["username"].(interface{})).(string))
			//ctx.Set("X-Request-User", (claims["username"].(interface{}).(string)))
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
	key := fmt.Sprintf("role_permission:%d", roleid)
	if exists, _ := config.CahceClient.Exists(context.Background(), key).Result(); exists < 1 {
		return false
	}
	exists, err := config.CahceClient.SIsMember(context.Background(), key, path).Result()
	if err != nil {
		return false
	}
	if exists {
		return true
	}
	// cache不存在就从数据库中加载
	if err := pkg.LoadPermissionToRedis(roleid); err != nil {
		return false
	}
	//// 查找role_permissions里是否有请求权限
	//	//if err := config.DB.Raw("SELECT p.name,p.group,p.`id`  from permissions p JOIN  role_permissions rp ON rp.permission_id = p.id where rp.role_id = ?", roleid).Scan(&ups).Error; err != nil {
	//	//	return false
	//	//}
	//	//for _, g := range ups {
	//	//	if g.Name == path {
	//	//		return true
	//	//	}
	//	//}
	exists, err = config.CahceClient.SIsMember(context.Background(), key, path).Result()
	if err != nil {
		return false
	}
	if exists {
		return true
	}
	return false
}
func formatRequestPath(path string) string {
	var perm string
	pats := strings.Split(path, "/")
	if len(pats) >= 4 {
		perm = strings.Join(pats[:4], "/")
	} else {
		perm = path
	}
	return perm
}
