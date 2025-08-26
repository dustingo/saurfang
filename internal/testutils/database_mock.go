package testutils

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MockDB 包含mock数据库的相关信息
type MockDB struct {
	DB   *gorm.DB
	Mock sqlmock.Sqlmock
	Conn *sql.DB
}

// SetupMockDB 创建并配置mock数据库
func SetupMockDB(t *testing.T) *MockDB {
	// 创建mock数据库连接
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)

	// 配置GORM
	dialector := mysql.New(mysql.Config{
		Conn:                      mockDB,
		SkipInitializeWithVersion: true,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)

	return &MockDB{
		DB:   db,
		Mock: mock,
		Conn: mockDB,
	}
}

// Close 关闭mock数据库连接
func (m *MockDB) Close() {
	m.Conn.Close()
}

// ExpectationsWereMet 验证所有mock期望都被满足
func (m *MockDB) ExpectationsWereMet(t *testing.T) {
	assert.NoError(t, m.Mock.ExpectationsWereMet())
}
