package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"saurfang/internal/config"
	"saurfang/internal/tools/pkg"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v3"
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
					return pkg.NewAppResponse(ctx, code, 1, http.StatusText(code), "", nil)
				}
				// 凭证校验通过,继续进行权限校验
				roleid, err := pkg.GetRoleOfUser(userid)
				if err != nil {
					return pkg.NewAppResponse(ctx, fiber.StatusUnauthorized, 1, err.Error(), "", nil)
				}
				//perm := formatRequestPath(requestPath)
				if hasPermission(roleid, perm) {
					ctx.Request().Header.Set("X-Request-User", strconv.Itoa(int(userid)))
					//ctx.Set("X-Request-User", (claims["username"].(interface{}).(string)))
					return ctx.Next()
				} else {
					return pkg.NewAppResponse(ctx, fiber.StatusUnauthorized, 1, "unauthorized", "", nil)
				}

			} else {
				// 凭证不存在，jwt也不存在
				return pkg.NewAppResponse(ctx, fiber.StatusUnauthorized, 1, "unauthorized", "", nil)
			}
		}
		token, err := jwt.Parse(ck, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_TOKEN_SECRET")), nil
		})
		if err != nil || !token.Valid {
			return pkg.NewAppResponse(ctx, fiber.StatusUnauthorized, 1, "unauthorized", "", nil)
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return pkg.NewAppResponse(ctx, fiber.StatusUnauthorized, 1, "unauthorized", "", nil)
		}
		role := claims["role"].(interface{})
		perm := formatRequestPath(requestPath)
		if hasPermission(uint(role.(float64)), perm) {
			ctx.Request().Header.Set("X-Request-User", (claims["username"].(interface{})).(string))
			//ctx.Set("X-Request-User", (claims["username"].(interface{}).(string)))
			return ctx.Next()
		} else {
			return pkg.NewAppResponse(ctx, fiber.StatusUnauthorized, 1, "unauthorized", "", nil)
		}
	}
}
func hasPermission(roleid uint, path string) bool {
	key := fmt.Sprintf("role_permission:%d", roleid)
	if exists, _ := config.CahceClient.Exists(context.Background(), key).Result(); exists < 1 {
		return false
	}
	// 检查缓存中是否存在该权限
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
	// 再次从缓存中查询,如果还是不存在,就返回错误
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
