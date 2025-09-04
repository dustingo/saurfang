package tools

import (
	"os"
	"saurfang/internal/config"
	"saurfang/internal/testutils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestUploadToOss 测试上传对象存储
// 不测试具体命令执行
// 不对上传函数做过多的拆分包装
func TestUploadToOss(t *testing.T) {
	mockDB := testutils.SetupMockDB(t)
	defer mockDB.Close()
	config.DB = mockDB.DB
	os.Setenv("SERVER_PACKAGE_DEST_PATH", "/tmp")
	os.Setenv("TESTING", "true")
	mockDB.Mock.ExpectQuery("select \\* from datasources where id = ?").
		WithArgs(1).
		WillReturnRows(mockDB.Mock.NewRows([]string{"id", "created_at", "updated_at", "access_key", "secret_key", "end_point", "region", "bucket", "path", "provider", "profile", "label"}).
			AddRow("1", time.Now(), time.Now(), "LTAI5t96tqD", "Y4xeXmAPVxA6VnYcCWj", "oss-cn-shanghai-internal.aliyuncs.com", "cn-shanghai", "catchfish-backup", "/cn", "Alibaba", "aliyun", "FISH-CN"))

	path, source, err := UploadToOss(1)
	assert.NoError(t, err)
	assert.Equal(t, "/cn", path)
	assert.Equal(t, "FISH-CN", source)
	assert.NoError(t, mockDB.Mock.ExpectationsWereMet())
}
