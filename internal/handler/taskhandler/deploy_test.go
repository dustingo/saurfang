package taskhandler_test

import (
	"bytes"
	"encoding/json"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http/httptest"
	"saurfang/internal/handler/taskhandler"
	"saurfang/internal/models/datasource"
	"saurfang/internal/models/task.go"
	"saurfang/internal/service/taskservice"
	"testing"
)

type MockDeployService struct {
	mock.Mock
	taskservice.DeployService
}

func (m *MockDeployService) ShowdsByID(id string) (*datasource.Datasources, error) {
	args := m.Called(id)
	return args.Get(0).(*datasource.Datasources), args.Error(1)
}
func (m *MockDeployService) CreateTask(payload task.DeployTaskPayload) error {
	args := m.Called(payload)
	return args.Error(0)
}

// 创建模拟的 DataSourceService
type MockDataSourceService struct {
	mock.Mock
}

func TestHandler_CreateDeployTask(t *testing.T) {
	// 初始化模拟对象
	mockDeployService := new(MockDeployService)
	mockDataSourceService := new(MockDataSourceService)

	//设置预期行为
	expectedDataSource := &datasource.Datasources{
		ID:    1,
		Label: "test-label",
	}
	mockDataSourceService.On("ShowdsByID", uint(1)).Return(expectedDataSource, nil)

	expectedTask := task.GameDeploymentTask{
		ID:          uint(1),
		SourceLabel: "test-label",
		Become:      1,
		BecomeUser:  "test-user",
		Comment:     "test-comment",
	}
	mockDeployService.On("CreateTask", &expectedTask).Return(nil)

	handler := taskhandler.DeployHandler{mockDeployService.DeployService}

	// 创建测试请求
	app := fiber.New()
	app.Post("/:id", handler.Handler_CreateDeployTask)

	payload := task.DeployTaskPayload{
		Become:     1,
		BecomeUser: "test-user",
		Comment:    "test-comment",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// 执行请求
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// 验证模拟调用
	mockDeployService.AssertExpectations(t)
	mockDataSourceService.AssertExpectations(t)
}
