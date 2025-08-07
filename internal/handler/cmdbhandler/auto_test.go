package cmdbhandler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"saurfang/internal/models/amis"
	"saurfang/internal/models/autosync"
	"saurfang/internal/repository/base"
	"saurfang/internal/tools/pkg"
)

// TestAutoSyncHandler_CreateAutoSyncConfig_Success 测试成功创建自动同步配置
func TestAutoSyncHandler_CreateAutoSyncConfig_Success(t *testing.T) {
	// 创建模拟数据库
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	// 创建GORM数据库连接
	dialector := mysql.New(mysql.Config{
		Conn:                      mockDB,
		SkipInitializeWithVersion: true,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)

	// 准备测试数据
	config := autosync.AutoSync{
		Cloud:     "阿里云",
		Label:     "test-label",
		Region:    "cn-hangzhou",
		Endpoint:  "https://ecs.cn-hangzhou.aliyuncs.com",
		GroupID:   "test-group",
		AccessKey: "test-access-key",
		SecretKey: "test-secret-key",
	}

	// 设置期望的SQL查询 - 插入自动同步配置
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `auto_syncs`").
		WithArgs(
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
			config.Cloud,
			config.Label,
			config.Region,
			config.Endpoint,
			config.GroupID,
			config.AccessKey,
			config.SecretKey,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// 创建处理器并注入mock数据库
	handler := &AutoSyncHandler{
		BaseGormRepository: base.BaseGormRepository[autosync.AutoSync]{
			DB: db,
		},
	}

	// 创建测试请求
	app := fiber.New()
	app.Post("/api/v1/cmdb/sync/create", func(c fiber.Ctx) error {
		return handler.Handler_CreateAutoSyncConfig(c)
	})

	body, _ := json.Marshal(config)
	req := httptest.NewRequest("POST", "/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	// 验证结果
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestAutoSyncHandler_CreateAutoSyncConfig_InvalidBody 测试无效请求体
func TestAutoSyncHandler_CreateAutoSyncConfig_InvalidBody(t *testing.T) {
	handler := &AutoSyncHandler{}

	app := fiber.New()
	app.Post("/config", func(c fiber.Ctx) error {
		return handler.Handler_CreateAutoSyncConfig(c)
	})

	req := httptest.NewRequest("POST", "/config", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestAutoSyncHandler_CreateAutoSyncConfig_DatabaseError 测试数据库错误
func TestAutoSyncHandler_CreateAutoSyncConfig_DatabaseError(t *testing.T) {
	config := autosync.AutoSync{
		Cloud:     "阿里云",
		Label:     "test-label",
		Region:    "cn-hangzhou",
		Endpoint:  "https://ecs.cn-hangzhou.aliyuncs.com",
		GroupID:   "test-group",
		AccessKey: "test-access-key",
		SecretKey: "test-secret-key",
	}

	handler := &AutoSyncHandler{}

	app := fiber.New()
	app.Post("/config", func(c fiber.Ctx) error {
		return handler.Handler_CreateAutoSyncConfig(c)
	})

	body, _ := json.Marshal(config)
	req := httptest.NewRequest("POST", "/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// TestAutoSyncHandler_ShowAutoSyncConfig_Success 测试成功显示配置列表
func TestAutoSyncHandler_ShowAutoSyncConfig_Success(t *testing.T) {
	// 创建模拟数据库
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	// 创建GORM数据库连接
	dialector := mysql.New(mysql.Config{
		Conn:                      mockDB,
		SkipInitializeWithVersion: true,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)

	// 准备测试数据
	expectedConfigs := []autosync.AutoSync{
		{
			ID:       1,
			Cloud:    "阿里云",
			Label:    "test-label-1",
			Region:   "cn-hangzhou",
			Endpoint: "https://ecs.cn-hangzhou.aliyuncs.com",
		},
		{
			ID:       2,
			Cloud:    "华为云",
			Label:    "test-label-2",
			Region:   "cn-north-4",
			Endpoint: "https://ecs.cn-north-4.myhuaweicloud.com",
		},
	}

	// 设置期望的SQL查询 - 查询所有自动同步配置
	mock.ExpectQuery("SELECT (.+) FROM `auto_syncs`").
		WillReturnRows(sqlmock.NewRows([]string{"id", "cloud", "label", "region", "endpoint"}).
			AddRow(expectedConfigs[0].ID, expectedConfigs[0].Cloud, expectedConfigs[0].Label, expectedConfigs[0].Region, expectedConfigs[0].Endpoint).
			AddRow(expectedConfigs[1].ID, expectedConfigs[1].Cloud, expectedConfigs[1].Label, expectedConfigs[1].Region, expectedConfigs[1].Endpoint))

	// 创建处理器并注入mock数据库
	handler := &AutoSyncHandler{
		BaseGormRepository: base.BaseGormRepository[autosync.AutoSync]{
			DB: db,
		},
	}

	// 创建测试请求
	app := fiber.New()
	app.Get("/configs", func(c fiber.Ctx) error {
		return handler.Handler_ShowAutoSyncConfig(c)
	})

	req := httptest.NewRequest("GET", "/configs", nil)
	resp, _ := app.Test(req)

	// 验证结果
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestAutoSyncHandler_UpdateAutoSyncConfig_Success 测试成功更新配置
func TestAutoSyncHandler_UpdateAutoSyncConfig_Success(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	dialector := mysql.New(mysql.Config{
		Conn:                      mockDB,
		SkipInitializeWithVersion: true,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)
	// 更新数据
	update := autosync.AutoSync{
		ID:        1,
		Cloud:     "阿里云",
		Label:     "new-label",
		Region:    "cn-shanghai",
		Endpoint:  "https://ecs.cn-shanghai.aliyuncs.com",
		GroupID:   "new-group-2",
		AccessKey: "new-access-key",
		SecretKey: "new-secret-key",
	}
	// GORM Save 会 UPDATE
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `auto_syncs` SET `created_at`=?,`updated_at`=?,`cloud`=?,`label`=?,`region`=?,`endpoint`=?,`group_id`=?,`access_key`=?,`secret_key`=? WHERE `id` = ?")).
		WithArgs(
			sqlmock.AnyArg(), // created_at - 使用 AnyArg 因为时间戳会变化
			sqlmock.AnyArg(), // updated_at - 使用 AnyArg 因为时间戳会变化
			update.Cloud,
			update.Label,
			update.Region,
			update.Endpoint,
			update.GroupID,
			update.AccessKey,
			update.SecretKey,
			update.ID, // id
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	handler := &AutoSyncHandler{
		BaseGormRepository: base.BaseGormRepository[autosync.AutoSync]{
			DB: db,
		},
	}

	app := fiber.New()
	app.Put("/sync/update/:id", func(c fiber.Ctx) error {
		return handler.Handler_UpdateAutoSyncConfig(c)
	})

	body, _ := json.Marshal(update)
	req := httptest.NewRequest("PUT", "/sync/update/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestAutoSyncHandler_UpdateAutoSyncConfig_InvalidID 测试无效ID
// statusCode 400
func TestAutoSyncHandler_UpdateAutoSyncConfig_InvalidID(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	dialector := mysql.New(mysql.Config{
		Conn:                      mockDB,
		SkipInitializeWithVersion: true,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)
	// 更新数据
	update := autosync.AutoSync{
		ID:        1,
		Cloud:     "阿里云",
		Label:     "new-label",
		Region:    "cn-shanghai",
		Endpoint:  "https://ecs.cn-shanghai.aliyuncs.com",
		GroupID:   "new-group-2",
		AccessKey: "new-access-key",
		SecretKey: "new-secret-key",
	}
	// GORM Save 会 UPDATE
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `auto_syncs` SET `created_at`=?,`updated_at`=?,`cloud`=?,`label`=?,`region`=?,`endpoint`=?,`group_id`=?,`access_key`=?,`secret_key`=? WHERE `id` = ?")).
		WithArgs(
			sqlmock.AnyArg(), // created_at - 使用 AnyArg 因为时间戳会变化
			sqlmock.AnyArg(), // updated_at - 使用 AnyArg 因为时间戳会变化
			update.Cloud,
			update.Label,
			update.Region,
			update.Endpoint,
			update.GroupID,
			update.AccessKey,
			update.SecretKey,
			update.ID, // id
		).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	handler := &AutoSyncHandler{
		BaseGormRepository: base.BaseGormRepository[autosync.AutoSync]{
			DB: db,
		},
	}

	app := fiber.New()
	app.Put("/sync/update/:id", func(c fiber.Ctx) error {
		return handler.Handler_UpdateAutoSyncConfig(c)
	})

	body, _ := json.Marshal(update)
	req := httptest.NewRequest("PUT", "/sync/update/2", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestAutoSyncHandler_DeleteAutoSyncConfig_Success 测试成功删除配置
func TestAutoSyncHandler_DeleteAutoSyncConfig_Success(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()
	dialector := mysql.New(mysql.Config{
		Conn:                      mockDB,
		SkipInitializeWithVersion: true,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `auto_syncs` WHERE `id` = ?").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	handler := &AutoSyncHandler{
		BaseGormRepository: base.BaseGormRepository[autosync.AutoSync]{
			DB: db,
		},
	}

	app := fiber.New()
	app.Delete("/api/v1/cmdb/sync/delete/:id", func(c fiber.Ctx) error {
		return handler.Handler_DeleteAutoSyncConfig(c)
	})

	req := httptest.NewRequest("DELETE", "/api/v1/cmdb/sync/delete/1", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestAutoSyncHandler_AutoSync_Success 测试成功自动同步
func TestAutoSyncHandler_AutoSync_Success(t *testing.T) {
	target := struct {
		Target string `json:"target"`
	}{
		Target: "阿里云-test-label",
	}

	handler := &AutoSyncHandler{}

	app := fiber.New()
	app.Post("/sync", func(c fiber.Ctx) error {
		return handler.Handler_AutoSync(c)
	})

	body, _ := json.Marshal(target)
	req := httptest.NewRequest("POST", "/sync", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	// 由于AutoSyncAliYunEcs需要真实的云服务调用，这里会返回错误
	// 在实际测试中，应该模拟云服务客户端
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// TestAutoSyncHandler_AutoSync_InvalidTarget 测试无效目标
func TestAutoSyncHandler_AutoSync_InvalidTarget(t *testing.T) {
	target := struct {
		Target string `json:"target"`
	}{
		Target: "invalid-target",
	}

	handler := &AutoSyncHandler{}

	app := fiber.New()
	app.Post("/sync", func(c fiber.Ctx) error {
		return handler.Handler_AutoSync(c)
	})

	body, _ := json.Marshal(target)
	req := httptest.NewRequest("POST", "/sync", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestAutoSyncHandler_AutoSyncConfigSelect_Success 测试成功获取配置选择器
func TestAutoSyncHandler_AutoSyncConfigSelect_Success(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()
	dialector := mysql.New(mysql.Config{
		Conn:                      mockDB,
		SkipInitializeWithVersion: true,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)
	// 原始数据
	syncConfigs := []autosync.AutoSync{
		{
			ID:       1,
			Cloud:    "阿里云",
			Label:    "test-label-1",
			Region:   "cn-hangzhou",
			Endpoint: "https://ecs.cn-hangzhou.aliyuncs.com",
		},
		{
			ID:       2,
			Cloud:    "华为云",
			Label:    "test-label-2",
			Region:   "cn-north-4",
			Endpoint: "https://ecs.cn-north-4.myhuaweicloud.com",
		},
	}
	// 接口返回的数据格式
	expectedResult := []amis.AmisOptionsString{
		{
			Label:      "test-label-1",
			SelectMode: "tree",
			Value:      "阿里云-test-label-1",
		},
		{
			Label:      "test-label-2",
			SelectMode: "tree",
			Value:      "华为云-test-label-2",
		},
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `auto_syncs`")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "cloud", "label", "region", "endpoint", "group_id", "access_key", "secret_key", "created_at", "updated_at"}).
			AddRow(syncConfigs[0].ID, syncConfigs[0].Cloud, syncConfigs[0].Label, syncConfigs[0].Region, syncConfigs[0].Endpoint, "", "", "", nil, nil).
			AddRow(syncConfigs[1].ID, syncConfigs[1].Cloud, syncConfigs[1].Label, syncConfigs[1].Region, syncConfigs[1].Endpoint, "", "", "", nil, nil))

	handler := &AutoSyncHandler{
		BaseGormRepository: base.BaseGormRepository[autosync.AutoSync]{
			DB: db,
		},
	}
	app := fiber.New()
	app.Get("/api/v1/cmdb/sync/select", func(c fiber.Ctx) error {
		return handler.Handler_AutoSyncConfigSelect(c)
	})

	req := httptest.NewRequest("GET", "/api/v1/cmdb/sync/select", nil)
	resp, _ := app.Test(req)
	body, _ := io.ReadAll(resp.Body)
	var result pkg.AppResponse
	err = json.Unmarshal(body, &result)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	options, ok := result.Data.(map[string]interface{})["options"]
	assert.True(t, ok)

	// Convert options to JSON and back to []amis.AmisOptionsString for comparison
	optionsBytes, err := json.Marshal(options)
	if !assert.NoError(t, err) {
		t.Errorf("failed to marshal options: %v", err)
	}

	expectBytes, err := json.Marshal(expectedResult)
	if !assert.NoError(t, err) {
		t.Errorf("failed to marshal expectedResult: %v", err)
	}

	assert.Equal(t, string(expectBytes), string(optionsBytes))
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestAutoSyncHandler_Constructor 测试构造函数
func TestAutoSyncHandler_Constructor(t *testing.T) {
	handler := &AutoSyncHandler{}
	assert.NotNil(t, handler)
	assert.IsType(t, &AutoSyncHandler{}, handler)
}
