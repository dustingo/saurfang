package userhandler_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"saurfang/internal/config"
	"saurfang/internal/handler/userhandler"
	"saurfang/internal/models/user"
	"saurfang/internal/repository/base"
	"saurfang/internal/testutils"
	"strconv"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// TestUserHandler_Handler_UserRegister_Success 测试成功注册
func TestUserHandler_Handler_UserRegister_Success(t *testing.T) {

	mockDB := testutils.SetupMockDB(t)
	defer mockDB.Close()

	// 配置GORM
	dialector := mysql.New(mysql.Config{
		Conn:                      mockDB.Conn,
		SkipInitializeWithVersion: true,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)

	// 设置全局DB配置
	config.DB = db

	// 创建UserHandler实例
	handler := &userhandler.UserHandler{
		BaseGormRepository: base.BaseGormRepository[user.User]{
			DB: db,
		},
	}
	// 准备测试数据
	// 正常数据
	payload := user.RegisterPayload{
		Username: "testuser",
		Password: "123456",
		Code:     "INVITE123",
	}
	// Mock邀请码查询 - 返回有效的邀请码
	mockDB.Mock.ExpectQuery("SELECT \\* FROM `invite_codes` WHERE code = \\? ORDER BY `invite_codes`.`id` LIMIT \\?").
		WithArgs(payload.Code, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "used"}).
			AddRow(1, "INVITE123", 0))

	// Mock用户创建查询 - 检查用户是否存在
	mockDB.Mock.ExpectQuery("SELECT \\* FROM `users` WHERE username = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(payload.Username, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// Mock用户创建插入 - 直接执行INSERT语句
	mockDB.Mock.ExpectExec("INSERT INTO users \\(username, password, created_at, updated_at\\) VALUES \\(\\?, \\?, NOW\\(\\), NOW\\(\\)\\)").
		WithArgs(payload.Username, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock获取新创建用户的查询
	mockDB.Mock.ExpectQuery("SELECT \\* FROM `users` WHERE username = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(payload.Username, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password"}).
			AddRow(1, payload.Username, "hashedpassword"))

	// Mock用户角色分配
	mockDB.Mock.ExpectExec("INSERT INTO user_roles \\(user_id, role_id\\) VALUES \\(\\?, \\?\\)").
		WithArgs(1, 4).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock邀请码状态更新
	mockDB.Mock.ExpectExec("UPDATE invite_codes SET used = 1 WHERE code = \\?").
		WithArgs(payload.Code).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 创建mock Fiber上下文
	app := fiber.New()
	app.Post("/api/v1/common/auth/register", handler.Handler_UserRegister)
	// c := app.AcquireCtx(&fasthttp.RequestCtx{})
	// defer app.ReleaseCtx(c)

	// 设置请求体
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/common/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// 执行测试
	resp, _ := app.Test(req)

	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// 验证所有mock期望都被满足
	assert.NoError(t, mockDB.Mock.ExpectationsWereMet())
}

// TestUserHandler_Handler_UserRegister_UsedInviteCode 测试使用已使用的邀请码注册
func TestUserHandler_Handler_UserRegister_UsedInviteCode(t *testing.T) {
	// 创建mock数据库
	mockDB := testutils.SetupMockDB(t)
	defer mockDB.Close()

	// 配置GORM
	dialector := mysql.New(mysql.Config{
		Conn:                      mockDB.Conn,
		SkipInitializeWithVersion: true,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)

	// 设置全局DB配置
	config.DB = db

	// 创建UserHandler实例
	handler := &userhandler.UserHandler{
		BaseGormRepository: base.BaseGormRepository[user.User]{
			DB: db,
		},
	}
	// 准备测试数据
	payload := user.RegisterPayload{
		Username: "testuser",
		Password: "123456",
		Code:     "INVITE123",
	}
	// Mock邀请码查询 - 返回无效的邀请码
	mockDB.Mock.ExpectQuery("SELECT \\* FROM `invite_codes` WHERE code = \\? ORDER BY `invite_codes`.`id` LIMIT \\?").
		WithArgs(payload.Code, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "used"}).
			AddRow(1, "INVITE123", 1))
	// 创建mock Fiber上下文
	app := fiber.New()
	app.Post("/api/v1/common/auth/register", handler.Handler_UserRegister)
	// 设置请求体
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/common/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// 执行测试
	resp, _ := app.Test(req)
	// 解析响应体
	var respBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&respBody)
	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, "invite code already used", respBody["message"])
}

// TestUserHandler_Handler_UserRegister_InvalidInviteCode 测试非法邀请码
func TestUserHandler_Handler_UserRegister_InvalidInviteCode(t *testing.T) {
	// 创建mock数据库
	mockDB := testutils.SetupMockDB(t)
	defer mockDB.Close()

	// 配置GORM
	dialector := mysql.New(mysql.Config{
		Conn:                      mockDB.Conn,
		SkipInitializeWithVersion: true,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)

	// 设置全局DB配置
	config.DB = db

	// 创建UserHandler实例
	handler := &userhandler.UserHandler{
		BaseGormRepository: base.BaseGormRepository[user.User]{
			DB: db,
		},
	}
	// 准备测试数据
	payload := user.RegisterPayload{
		Username: "testuser",
		Password: "123456",
		Code:     "INVITE123",
	}
	// Mock邀请码查询 - 找不到邀请码
	mockDB.Mock.ExpectQuery("SELECT \\* FROM `invite_codes` WHERE code = \\? ORDER BY `invite_codes`.`id` LIMIT \\?").
		WithArgs(payload.Code, 1).WillReturnError(gorm.ErrRecordNotFound)
	// 创建mock Fiber上下文
	app := fiber.New()
	app.Post("/api/v1/common/auth/register", handler.Handler_UserRegister)
	// 设置请求体
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/common/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// 执行测试
	resp, _ := app.Test(req)
	// 解析响应体
	var respBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&respBody)
	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, "invite code not found", respBody["message"])
	// 验证所有mock期望都被满足
	assert.NoError(t, mockDB.Mock.ExpectationsWereMet())
}

// TestUserHandler_Handler_UserRegister_EmptyInviteCode 测试空邀请码
func TestUserHandler_Handler_UserRegister_EmptyInviteCode(t *testing.T) {
	// 创建mock数据库
	mockDB := testutils.SetupMockDB(t)
	defer mockDB.Close()

	// 配置GORM
	dialector := mysql.New(mysql.Config{
		Conn:                      mockDB.Conn,
		SkipInitializeWithVersion: true,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)

	// 设置全局DB配置
	config.DB = db

	// 创建UserHandler实例
	handler := &userhandler.UserHandler{
		BaseGormRepository: base.BaseGormRepository[user.User]{
			DB: db,
		},
	}
	// 准备测试数据
	payload := user.RegisterPayload{
		Username: "testuser",
		Password: "123456",
	}
	app := fiber.New()
	app.Post("/api/v1/common/auth/register", handler.Handler_UserRegister)
	// 设置请求体
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/common/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// 执行测试
	resp, _ := app.Test(req)
	// 解析响应体
	var respBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&respBody)
	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, "invite code is required", respBody["message"])
	// 验证所有mock期望都被满足
	assert.NoError(t, mockDB.Mock.ExpectationsWereMet())
}
func TestUserHandler_Handler_UserRegister_InvalidUsername(t *testing.T) {
	// 创建mock数据库
	mockDB := testutils.SetupMockDB(t)
	defer mockDB.Close()

	// 配置GORM
	dialector := mysql.New(mysql.Config{
		Conn:                      mockDB.Conn,
		SkipInitializeWithVersion: true,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)

	// 设置全局DB配置
	config.DB = db

	// 创建UserHandler实例
	handler := &userhandler.UserHandler{
		BaseGormRepository: base.BaseGormRepository[user.User]{
			DB: db,
		},
	}
	// 准备测试数据
	payload := user.RegisterPayload{
		Username: "te",
		Password: "123456",
		Code:     "INVITE123",
	}
	app := fiber.New()
	app.Post("/api/v1/common/auth/register", handler.Handler_UserRegister)
	// 设置请求体
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/common/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// 执行测试
	resp, _ := app.Test(req)
	// 解析响应体
	var respBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&respBody)
	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, "username validation failed", respBody["message"])
}

// TestUserHandler_Handler_UserLogin_Success 测试登录成功
func TestUserHandler_Handler_UserLogin_Success(t *testing.T) {
	// 创建mock数据库
	mockDB := testutils.SetupMockDB(t)
	defer mockDB.Close()

	// 配置GORM
	dialector := mysql.New(mysql.Config{
		Conn:                      mockDB.Conn,
		SkipInitializeWithVersion: true,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)
	// 设置全局DB配置
	config.DB = db

	// 创建UserHandler实例
	handler := &userhandler.UserHandler{
		BaseGormRepository: base.BaseGormRepository[user.User]{
			DB: db,
		},
	}
	// 准备测试数据
	payload := user.LoginPayload{
		Username: "testuser",
		Password: "123456",
	}
	os.Setenv("JWT_TOKEN_EXP", "3600")
	os.Setenv("JWT_TOKEN_SECRET", "test-secret-key")
	defer func() {
		os.Unsetenv("JWT_TOKEN_EXP")
		os.Unsetenv("JWT_TOKEN_SECRET")
	}()
	expireTime, _ := strconv.Atoi(os.Getenv("JWT_TOKEN_EXP"))
	fixedTime := time.Date(2025, 8, 26, 18, 20, 0, 0, time.Local)
	//测试jwt token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       1,
		"username": payload.Username,
		"role":     1,
		"iat":      fixedTime.Unix(),
		"exp":      fixedTime.Add(time.Duration(expireTime) * time.Second).Unix(),
	})
	tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_TOKEN_SECRET")))
	//解析token
	newToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_TOKEN_SECRET")), nil
	})
	assert.NoError(t, err)
	// 验证token
	claims, ok := newToken.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, "testuser", claims["username"])
	assert.Equal(t, float64(1), claims["role"])
	assert.Equal(t, float64(1), claims["id"])
	assert.Equal(t, float64(fixedTime.Unix()), claims["iat"])
	assert.Equal(t, float64(fixedTime.Add(time.Duration(expireTime)*time.Second).Unix()), claims["exp"])
	// 继续业务逻辑
	// 查询用户
	mockDB.Mock.ExpectQuery("SELECT \\* FROM `users` WHERE username = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(payload.Username, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "token", "code"}).
			AddRow(1, "testuser", "$2a$10$2jse9R8IfgGoLbjOAg..1uJ9jBn0vY3LS/Nl7fnHODFWWLMDij2de", "token123", "INVITE123"))

	// Mock RoleOfUser查询 - tools.RoleOfUser函数会调用这个查询
	mockDB.Mock.ExpectQuery("SELECT \\* FROM `user_roles` WHERE user_id = \\? ORDER BY `user_roles`.`role_id` LIMIT \\?").
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "role_id"}).
			AddRow(1, 1))

	// 执行测试
	app := fiber.New()
	app.Post("/api/v1/common/auth/login", handler.Handler_UserLogin)
	resp, err := testutils.CreateHTTPTestRequest(app, "POST", "/api/v1/common/auth/login", payload)
	assert.NoError(t, err)
	// 解析响应体
	var respBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&respBody)
	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "login success", respBody["message"])
}

// TestUserHandler_Handler_UserLogin_InvalidPassword 测试登录失败-密码错误
func TestUserHandler_Handler_UserLogin_InvalidPassword(t *testing.T) {
	// 创建mock数据库
	mockDB := testutils.SetupMockDB(t)
	defer mockDB.Close()

	// 配置GORM
	dialector := mysql.New(mysql.Config{
		Conn:                      mockDB.Conn,
		SkipInitializeWithVersion: true,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)
	// 设置全局DB配置
	config.DB = db

	// 创建UserHandler实例
	handler := &userhandler.UserHandler{
		BaseGormRepository: base.BaseGormRepository[user.User]{
			DB: db,
		},
	}
	// 准备测试数据
	payload := user.LoginPayload{
		Username: "testuser",
		Password: "1234567",
	}
	os.Setenv("JWT_TOKEN_EXP", "3600")
	os.Setenv("JWT_TOKEN_SECRET", "test-secret-key")
	defer func() {
		os.Unsetenv("JWT_TOKEN_EXP")
		os.Unsetenv("JWT_TOKEN_SECRET")
	}()
	// 继续业务逻辑
	// 查询用户
	mockDB.Mock.ExpectQuery("SELECT \\* FROM `users` WHERE username = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(payload.Username, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "token", "code"}).
			AddRow(1, "testuser", "$2a$10$2jse9R8IfgGoLbjOAg..1uJ9jBn0vY3LS/Nl7fnHODFWWLMDij2de", "token123", "INVITE123"))

	// Mock RoleOfUser查询 - tools.RoleOfUser函数会调用这个查询
	mockDB.Mock.ExpectQuery("SELECT \\* FROM `user_roles` WHERE user_id = \\? ORDER BY `user_roles`.`role_id` LIMIT \\?").
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "role_id"}).
			AddRow(1, 1))

	// 执行测试
	app := fiber.New()
	app.Post("/api/v1/common/auth/login", handler.Handler_UserLogin)
	resp, err := testutils.CreateHTTPTestRequest(app, "POST", "/api/v1/common/auth/login", payload)
	assert.NoError(t, err)
	// 解析响应体
	var respBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&respBody)
	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	assert.Equal(t, "password is wrong", respBody["message"])
}

// TestUserHandler_Handler_UserLogin_InvalidUsername 测试登录失败-用户名错误
func TestUserHandler_Handler_UserLogin_InvalidUsername(t *testing.T) {
	// 创建mock数据库
	mockDB := testutils.SetupMockDB(t)
	defer mockDB.Close()

	// 配置GORM
	dialector := mysql.New(mysql.Config{
		Conn:                      mockDB.Conn,
		SkipInitializeWithVersion: true,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)
	// 设置全局DB配置
	config.DB = db

	// 创建UserHandler实例
	handler := &userhandler.UserHandler{
		BaseGormRepository: base.BaseGormRepository[user.User]{
			DB: db,
		},
	}
	// 准备测试数据
	payload := user.LoginPayload{
		Username: "test",
		Password: "123456",
	}
	os.Setenv("JWT_TOKEN_EXP", "3600")
	os.Setenv("JWT_TOKEN_SECRET", "test-secret-key")
	defer func() {
		os.Unsetenv("JWT_TOKEN_EXP")
		os.Unsetenv("JWT_TOKEN_SECRET")
	}()
	// 继续业务逻辑
	// 查询用户
	mockDB.Mock.ExpectQuery("SELECT \\* FROM `users` WHERE username = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(payload.Username, 1).WillReturnError(gorm.ErrRecordNotFound)

	// Mock RoleOfUser查询 - tools.RoleOfUser函数会调用这个查询
	mockDB.Mock.ExpectQuery("SELECT \\* FROM `user_roles` WHERE user_id = \\? ORDER BY `user_roles`.`role_id` LIMIT \\?").
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "role_id"}).
			AddRow(1, 1))

	// 执行测试
	app := fiber.New()
	app.Post("/api/v1/common/auth/login", handler.Handler_UserLogin)
	resp, err := testutils.CreateHTTPTestRequest(app, "POST", "/api/v1/common/auth/login", payload)
	assert.NoError(t, err)
	// 解析响应体
	var respBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&respBody)
	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	assert.Equal(t, "user not exist", respBody["message"])
}
