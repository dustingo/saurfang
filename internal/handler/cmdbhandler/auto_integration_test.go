package cmdbhandler

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"saurfang/internal/models/autosync"
	"saurfang/internal/models/gamehost"
	"saurfang/internal/repository/base"
)

// TestAutoSyncHandler_AutoSyncAliYunEcs_Integration 阿里云ECS同步集成测试
func TestAutoSyncHandler_AutoSyncAliYunEcs_Integration(t *testing.T) {
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

	// 创建处理器
	handler := &AutoSyncHandler{
		BaseGormRepository: base.BaseGormRepository[autosync.AutoSync]{
			DB: db,
		},
	}

	// 设置期望的SQL查询
	// 1. 查询自动同步配置
	mock.ExpectQuery("SELECT (.+) FROM `auto_syncs` WHERE label = (.+) ORDER BY `auto_syncs`.`id` LIMIT 1").
		WithArgs("test-label").
		WillReturnRows(sqlmock.NewRows([]string{"id", "cloud", "label", "region", "endpoint", "group_id", "access_key", "secret_key"}).
			AddRow(1, "阿里云", "test-label", "cn-hangzhou", "https://ecs.cn-hangzhou.aliyuncs.com", "test-group", "test-access-key", "test-secret-key"))

	// 2. 查询本地主机数据
	mock.ExpectQuery("select \\* from hosts").
		WillReturnRows(sqlmock.NewRows([]string{"id", "instance_id", "hostname", "public_ip", "private_ip", "cpu", "memory", "os_name"}).
			AddRow(1, "i-test123", "test-host", "1.2.3.4", "10.0.0.1", "2", "4", "CentOS"))

	// 3. 执行更新SQL（模拟有变更的情况）
	mock.ExpectExec("UPDATE hosts SET hostname = (.+), public_ip = (.+), private_ip = (.+), cpu = (.+), memory = (.+), os_name = (.+) WHERE instance_id = (.+)").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 执行测试
	err = handler.AutoSyncAliYunEcs("test-label")

	// 验证结果
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestAutoSyncHandler_AutoSyncHuaweiECS_Integration 华为云ECS同步集成测试
func TestAutoSyncHandler_AutoSyncHuaweiECS_Integration(t *testing.T) {
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

	// 创建处理器
	handler := &AutoSyncHandler{
		BaseGormRepository: base.BaseGormRepository[autosync.AutoSync]{
			DB: db,
		},
	}

	// 设置期望的SQL查询
	// 1. 查询自动同步配置
	mock.ExpectQuery("SELECT (.+) FROM `auto_syncs` WHERE label = (.+) ORDER BY `auto_syncs`.`id` LIMIT 1").
		WithArgs("test-label").
		WillReturnRows(sqlmock.NewRows([]string{"id", "cloud", "label", "region", "endpoint", "group_id", "access_key", "secret_key"}).
			AddRow(1, "华为云", "test-label", "cn-north-4", "https://ecs.cn-north-4.myhuaweicloud.com", "test-group", "test-access-key", "test-secret-key"))

	// 2. 查询本地主机数据
	mock.ExpectQuery("select \\* from hosts").
		WillReturnRows(sqlmock.NewRows([]string{"id", "instance_id", "hostname", "public_ip", "private_ip", "cpu", "memory", "os_name"}).
			AddRow(1, "i-test456", "test-host", "1.2.3.5", "10.0.0.2", "4", "8", "Ubuntu"))

	// 3. 执行插入SQL（模拟新增的情况）
	mock.ExpectExec("INSERT INTO hosts \\(created_at, instance_id, hostname, public_ip, private_ip, cpu, memory, os_name, instance_type, port\\) VALUES\\((.+), (.+), (.+), (.+), (.+), (.+), (.+), (.+), (.+), (.+)\\)").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 执行测试
	err = handler.AutoSyncHuaweiECS("test-label")

	// 验证结果
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestAutoSyncHandler_DatabaseTransaction_Integration 测试数据库事务
func TestAutoSyncHandler_DatabaseTransaction_Integration(t *testing.T) {
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

	// 创建处理器
	handler := &AutoSyncHandler{
		BaseGormRepository: base.BaseGormRepository[autosync.AutoSync]{
			DB: db,
		},
	}

	// 设置期望的SQL查询
	// 1. 查询自动同步配置
	mock.ExpectQuery("SELECT (.+) FROM `auto_syncs` WHERE label = (.+) ORDER BY `auto_syncs`.`id` LIMIT 1").
		WithArgs("test-label").
		WillReturnRows(sqlmock.NewRows([]string{"id", "cloud", "label", "region", "endpoint", "group_id", "access_key", "secret_key"}).
			AddRow(1, "阿里云", "test-label", "cn-hangzhou", "https://ecs.cn-hangzhou.aliyuncs.com", "test-group", "test-access-key", "test-secret-key"))

	// 2. 查询本地主机数据
	mock.ExpectQuery("select \\* from hosts").
		WillReturnRows(sqlmock.NewRows([]string{"id", "instance_id", "hostname", "public_ip", "private_ip", "cpu", "memory", "os_name"}))

	// 3. 开始事务
	mock.ExpectBegin()

	// 4. 执行SQL（模拟成功）
	mock.ExpectExec("UPDATE hosts SET hostname = (.+), public_ip = (.+), private_ip = (.+), cpu = (.+), memory = (.+), os_name = (.+) WHERE instance_id = (.+)").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 5. 提交事务
	mock.ExpectCommit()

	// 执行测试
	err = handler.AutoSyncAliYunEcs("test-label")

	// 验证结果
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestAutoSyncHandler_DatabaseTransactionRollback_Integration 测试数据库事务回滚
func TestAutoSyncHandler_DatabaseTransactionRollback_Integration(t *testing.T) {
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

	// 创建处理器
	handler := &AutoSyncHandler{
		BaseGormRepository: base.BaseGormRepository[autosync.AutoSync]{
			DB: db,
		},
	}

	// 设置期望的SQL查询
	// 1. 查询自动同步配置
	mock.ExpectQuery("SELECT (.+) FROM `auto_syncs` WHERE label = (.+) ORDER BY `auto_syncs`.`id` LIMIT 1").
		WithArgs("test-label").
		WillReturnRows(sqlmock.NewRows([]string{"id", "cloud", "label", "region", "endpoint", "group_id", "access_key", "secret_key"}).
			AddRow(1, "阿里云", "test-label", "cn-hangzhou", "https://ecs.cn-hangzhou.aliyuncs.com", "test-group", "test-access-key", "test-secret-key"))

	// 2. 查询本地主机数据
	mock.ExpectQuery("select \\* from hosts").
		WillReturnRows(sqlmock.NewRows([]string{"id", "instance_id", "hostname", "public_ip", "private_ip", "cpu", "memory", "os_name"}))

	// 3. 开始事务
	mock.ExpectBegin()

	// 4. 执行SQL（模拟失败）
	mock.ExpectExec("UPDATE hosts SET hostname = (.+), public_ip = (.+), private_ip = (.+), cpu = (.+), memory = (.+), os_name = (.+) WHERE instance_id = (.+)").
		WillReturnError(sql.ErrConnDone)

	// 5. 回滚事务
	mock.ExpectRollback()

	// 执行测试
	err = handler.AutoSyncAliYunEcs("test-label")

	// 验证结果
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestAutoSyncHandler_DataHashComparison_Integration 测试数据哈希比较
func TestAutoSyncHandler_DataHashComparison_Integration(t *testing.T) {
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

	// 创建处理器
	handler := &AutoSyncHandler{
		BaseGormRepository: base.BaseGormRepository[autosync.AutoSync]{
			DB: db,
		},
	}

	// 准备测试数据
	localHosts := []gamehost.Hosts{
		{
			InstanceID: "i-test123",
			Hostname:   "test-host-1",
			PublicIP:   "1.2.3.4",
			PrivateIP:  "10.0.0.1",
			CPU:        "2",
			Memory:     "4",
		},
		{
			InstanceID: "i-test456",
			Hostname:   "test-host-2",
			PublicIP:   "1.2.3.5",
			PrivateIP:  "10.0.0.2",
			CPU:        "4",
			Memory:     "8",
		},
	}

	// 设置期望的SQL查询
	// 1. 查询自动同步配置
	mock.ExpectQuery("SELECT (.+) FROM `auto_syncs` WHERE label = (.+) ORDER BY `auto_syncs`.`id` LIMIT 1").
		WithArgs("test-label").
		WillReturnRows(sqlmock.NewRows([]string{"id", "cloud", "label", "region", "endpoint", "group_id", "access_key", "secret_key"}).
			AddRow(1, "阿里云", "test-label", "cn-hangzhou", "https://ecs.cn-hangzhou.aliyuncs.com", "test-group", "test-access-key", "test-secret-key"))

	// 2. 查询本地主机数据
	rows := sqlmock.NewRows([]string{"id", "instance_id", "hostname", "public_ip", "private_ip", "cpu", "memory", "os_name"})
	for _, host := range localHosts {
		rows.AddRow(1, host.InstanceID, host.Hostname, host.PublicIP, host.PrivateIP, host.CPU, host.Memory, "CentOS")
	}
	mock.ExpectQuery("select \\* from hosts").WillReturnRows(rows)

	// 3. 开始事务
	mock.ExpectBegin()

	// 4. 执行SQL（模拟有变更的情况）
	mock.ExpectExec("UPDATE hosts SET hostname = (.+), public_ip = (.+), private_ip = (.+), cpu = (.+), memory = (.+), os_name = (.+) WHERE instance_id = (.+)").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 5. 提交事务
	mock.ExpectCommit()

	// 执行测试
	err = handler.AutoSyncAliYunEcs("test-label")

	// 验证结果
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestAutoSyncHandler_ErrorHandling_Integration 测试错误处理
func TestAutoSyncHandler_ErrorHandling_Integration(t *testing.T) {
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

	// 创建处理器
	handler := &AutoSyncHandler{
		BaseGormRepository: base.BaseGormRepository[autosync.AutoSync]{
			DB: db,
		},
	}

	// 设置期望的SQL查询 - 配置不存在
	mock.ExpectQuery("SELECT (.+) FROM `auto_syncs` WHERE label = (.+) ORDER BY `auto_syncs`.`id` LIMIT 1").
		WithArgs("non-existent-label").
		WillReturnError(sql.ErrNoRows)

	// 执行测试
	err = handler.AutoSyncAliYunEcs("non-existent-label")

	// 验证结果
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 性能测试
func BenchmarkAutoSyncHandler_AutoSyncAliYunEcs(b *testing.B) {
	// 创建模拟数据库
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		b.Fatal(err)
	}
	defer mockDB.Close()

	// 创建GORM数据库连接
	dialector := mysql.New(mysql.Config{
		Conn:                      mockDB,
		SkipInitializeWithVersion: true,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		b.Fatal(err)
	}

	// 创建处理器
	handler := &AutoSyncHandler{
		BaseGormRepository: base.BaseGormRepository[autosync.AutoSync]{
			DB: db,
		},
	}

	// 设置期望的SQL查询
	mock.ExpectQuery("SELECT (.+) FROM `auto_syncs` WHERE label = (.+) ORDER BY `auto_syncs`.`id` LIMIT 1").
		WithArgs("test-label").
		WillReturnRows(sqlmock.NewRows([]string{"id", "cloud", "label", "region", "endpoint", "group_id", "access_key", "secret_key"}).
			AddRow(1, "阿里云", "test-label", "cn-hangzhou", "https://ecs.cn-hangzhou.aliyuncs.com", "test-group", "test-access-key", "test-secret-key"))

	mock.ExpectQuery("select \\* from hosts").
		WillReturnRows(sqlmock.NewRows([]string{"id", "instance_id", "hostname", "public_ip", "private_ip", "cpu", "memory", "os_name"}))

	mock.ExpectBegin()
	mock.ExpectCommit()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.AutoSyncAliYunEcs("test-label")
	}
}
