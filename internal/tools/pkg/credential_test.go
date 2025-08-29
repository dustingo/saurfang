package pkg_test

import (
	"fmt"
	"regexp"
	"saurfang/internal/config"
	"saurfang/internal/testutils"
	"saurfang/internal/tools/pkg"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestGenerateSecureRandomStringForTest 随机数测试
func TestGenerateSecureRandomStringForTest(t *testing.T) {
	t.Run("正常长度生成", func(t *testing.T) {
		lengths := []int{1, 5, 10, 20, 40, 100}
		for _, length := range lengths {
			str, err := pkg.GenerateSecureRandomStringForTest(length)
			assert.NoError(t, err)
			assert.Equal(t, length, len(str))
			// 检查字符串是否只包含有效字符
			for _, char := range str {
				assert.Contains(t, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789", string(char))
			}
		}
	})
	t.Run("零长度生成", func(t *testing.T) {
		str, err := pkg.GenerateSecureRandomStringForTest(0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "length must be greater than 0")
		assert.Equal(t, "", str)
	})
	t.Run("随机性验证", func(t *testing.T) {
		length := 10
		results := make(map[string]bool)
		for i := 0; i < 100; i++ {
			str, err := pkg.GenerateSecureRandomStringForTest(length)
			assert.NoError(t, err)
			results[str] = true
		}
		assert.Equal(t, 100, len(results))
	})
}

// TestVerifySignature 校验签名
func TestVerifySignature(t *testing.T) {
	mockDB := testutils.SetupMockDB(t)
	defer mockDB.Close()
	config.DB = mockDB.DB
	accessKey := "EMz9pu2jrgGXjzd21O8Q"
	secretKey := "EptyehZpagEtfrFkgRr2AQ2"
	method := "GET"
	path := "/api/v1/dashboard"
	// 使用当前时间戳
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	// 生成正确的签名
	signature, _ := pkg.GenerateSignatureForTest(accessKey, secretKey, method, path, timestamp)
	t.Run("校验签名", func(t *testing.T) {
		mockDB.Mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `user_credentials` WHERE access_key = ?")).
			WithArgs(accessKey).
			WillReturnRows(mockDB.Mock.NewRows([]string{"id", "access_key", "secret_key", "user_id", "status"}).
				AddRow(1, accessKey, secretKey, 1, "active"))
		userID, code, ok := pkg.VerifySignature(accessKey, signature, method, path, timestamp)
		assert.Equal(t, uint(1), userID)
		assert.Equal(t, 200, code)
		assert.Equal(t, true, ok)
		assert.NoError(t, mockDB.Mock.ExpectationsWereMet())
	})
}
