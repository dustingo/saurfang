package testutils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/gofiber/fiber/v3"
)

// CreateHTTPTestRequest 创建HTTP测试请求
func CreateHTTPTestRequest(app *fiber.App, method, url string, body interface{}) (*http.Response, error) {
	var req *http.Request

	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		req = httptest.NewRequest(method, url, bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, url, nil)
	}

	resp, err := app.Test(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
