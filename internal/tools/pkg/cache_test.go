package pkg_test

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"saurfang/internal/config"
	"saurfang/internal/models/notify"
	"saurfang/internal/testutils"
	"saurfang/internal/tools/pkg"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

// TestWarmUpCache 测试WarmUpCache初始化权限缓存
func TestWarmUpCache(t *testing.T) {
	mockDB := testutils.SetupMockDB(t)
	defer mockDB.Close()
	rdb, rdbmock := redismock.NewClientMock()
	defer rdb.Close()
	query := "SELECT r.id,  rp.permission_id, p.name  FROM roles r JOIN role_permissions rp ON r.id  = rp.role_id JOIN permissions p ON rp.permission_id = p.id order by id"
	mockDB.Mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(mockDB.Mock.NewRows([]string{"id", "permission_id", "name"}).AddRows([]driver.Value{1, "1", "/api/v1/cmdb"}, []driver.Value{1, "1", "/api/v1/credential"}, []driver.Value{2, "1", "/api/v1/cmdb"}))
	config.DB = mockDB.DB
	config.CahceClient = rdb

	// Del 操作期望 - 按照rps数组的顺序执行
	rdbmock.ExpectDel("role_permission:1").SetVal(1)
	rdbmock.ExpectDel("role_permission:1").SetVal(1)
	rdbmock.ExpectDel("role_permission:2").SetVal(1)

	// SAdd 操作期望 - 按照实际执行顺序
	rdbmock.ExpectSAdd(fmt.Sprintf("role_permission:%d", 1), "/api/v1/cmdb").SetVal(1)
	rdbmock.ExpectSAdd(fmt.Sprintf("role_permission:%d", 1), "/api/v1/credential").SetVal(1)
	rdbmock.ExpectSAdd(fmt.Sprintf("role_permission:%d", 2), "/api/v1/cmdb").SetVal(1)

	// Expire 操作期望 - 按照keys数组的顺序（包含重复）
	rdbmock.ExpectExpire(fmt.Sprintf("role_permission:%d", 1), 24*time.Hour).SetVal(true)
	rdbmock.ExpectExpire(fmt.Sprintf("role_permission:%d", 1), 24*time.Hour).SetVal(true)
	rdbmock.ExpectExpire(fmt.Sprintf("role_permission:%d", 2), 24*time.Hour).SetVal(true)

	err := pkg.WarmUpCache()
	assert.NoError(t, err)
	mockDB.ExpectationsWereMet(t)
	assert.NoError(t, rdbmock.ExpectationsWereMet())
}

// TestLoadPermissionToRedis 测试LoadPermissionToRedis加载指定用户权限到Redis
func TestLoadPermissionToRedis(t *testing.T) {
	mockDB := testutils.SetupMockDB(t)
	defer mockDB.Close()
	rdb, rdbmock := redismock.NewClientMock()
	defer rdb.Close()
	config.DB = mockDB.DB
	config.CahceClient = rdb
	// sql
	roleid := uint(1)
	query := fmt.Sprintf("SELECT r.id,  rp.permission_id, p.name  FROM roles r JOIN role_permissions rp ON r.id  = rp.role_id JOIN permissions p ON rp.permission_id = p.id WHERE r.id= %d order by permission_id", roleid)
	mockDB.Mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(mockDB.Mock.NewRows([]string{"id", "permission_id", "name"}).AddRows([]driver.Value{1, "1", "/api/v1/cmdb"}, []driver.Value{1, "2", "/api/v1/common"}, []driver.Value{1, "3", "/api/v1/credential"}))
	key := "role_permission:1"
	rdbmock.ExpectSAdd(key, "/api/v1/cmdb").SetVal(1)
	rdbmock.ExpectSAdd(key, "/api/v1/common").SetVal(1)
	rdbmock.ExpectSAdd(key, "/api/v1/credential").SetVal(1)
	rdbmock.ExpectExpire(key, 24*time.Hour).SetVal(true)
	err := pkg.LoadPermissionToRedis(roleid)
	assert.NoError(t, err)
	mockDB.ExpectationsWereMet(t)
	assert.NoError(t, rdbmock.ExpectationsWereMet())
}

// TestWarmUpNotifyCache 测试WarmUpNotifyCache初始化通知缓存
func TestWarmUpNotifyCache(t *testing.T) {
	mockDB := testutils.SetupMockDB(t)
	defer mockDB.Close()
	rdb, rdbmock := redismock.NewClientMock()
	defer rdb.Close()
	config.DB = mockDB.DB
	config.CahceClient = rdb

	// 期望查询notify_subscribes表
	mockDB.Mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `notify_subscribes`")).
		WillReturnRows(
			mockDB.Mock.NewRows(
				[]string{"id", "user_id", "event_type", "notify_config_id", "status", "created_at", "updated_at"}).
				AddRows(
					[]driver.Value{1, 1, "gameops,gamedeploy,customjob,cronjob,upload", 1, "0", time.Now(), time.Now()},
					[]driver.Value{3, 1, "upload", 2, "0", time.Now(), time.Now()},
					[]driver.Value{4, 1, "gameops,gamedeploy", 3, "0", time.Now(), time.Now()}))

	// 期望Keys操作获取现有缓存键
	pattern := fmt.Sprintf("%s:*", notify.SubscribeKey)
	rdbmock.ExpectKeys(pattern).SetVal([]string{"notify_subscribe:1", "notify_subscribe:3", "notify_subscribe:4"})

	// 期望Del操作清理旧缓存
	rdbmock.ExpectDel("notify_subscribe:1").SetVal(1)
	rdbmock.ExpectDel("notify_subscribe:3").SetVal(1)
	rdbmock.ExpectDel("notify_subscribe:4").SetVal(1)

	// 存储订阅详情 - 按照subs数组的顺序
	rdbmock.ExpectSet("notify_subscribe:detail:1", []byte("{\"event_type\":\"gameops,gamedeploy,customjob,cronjob,upload\",\"notify_config_id\":1,\"status\":\"0\",\"user_id\":1}"), 24*time.Hour).SetVal("OK")

	rdbmock.ExpectSet("notify_subscribe:detail:3", []byte("{\"event_type\":\"upload\",\"notify_config_id\":2,\"status\":\"0\",\"user_id\":1}"), 24*time.Hour).SetVal("OK")

	rdbmock.ExpectSet("notify_subscribe:detail:4", []byte("{\"event_type\":\"gameops,gamedeploy\",\"notify_config_id\":3,\"status\":\"0\",\"user_id\":1}"), 24*time.Hour).SetVal("OK")

	// 期望查询notify_configs表
	mockDB.Mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `notify_configs`")).
		WillReturnRows(mockDB.Mock.NewRows([]string{"id", "name", "channel", "config", "status", "created_at", "updated_at"}).AddRows(
			[]driver.Value{1, "dingding", "dingtalk", " {\"token\": \"d23e1663201b04198f1c\", \"secret\": \"SEC78ec9c11fe0372d\"}", 0, time.Now(), time.Now()},
			[]driver.Value{2, "email", "email", "{\"to\": [\"doumaoxin@360.cn\", \"514838728@qq.com\"]}", 0, time.Now(), time.Now()},
			[]driver.Value{3, "lark", "lark", "{\"webhook\": \"https://open.feishu.cn/open-apis/bot/v2/hook/c1506ec0\"}", 0, time.Now(), time.Now()},
		))

	// 存储配置详情 - 按照configs数组的顺序，JSON字段顺序需要与实际序列化顺序一致
	rdbmock.ExpectSet("notify_config:detail:1", []byte("{\"channel\":\"dingtalk\",\"config\":{\"token\":\"d23e1663201b04198f1c\",\"secret\":\"SEC78ec9c11fe0372d\"},\"name\":\"dingding\",\"status\":\"0\"}"), 24*time.Hour).SetVal("OK")
	rdbmock.ExpectSet("notify_config:detail:2", []byte("{\"channel\":\"email\",\"config\":{\"to\":[\"doumaoxin@360.cn\",\"514838728@qq.com\"]},\"name\":\"email\",\"status\":\"0\"}"), 24*time.Hour).SetVal("OK")
	rdbmock.ExpectSet("notify_config:detail:3", []byte("{\"channel\":\"lark\",\"config\":{\"webhook\":\"https://open.feishu.cn/open-apis/bot/v2/hook/c1506ec0\"},\"name\":\"lark\",\"status\":\"0\"}"), 24*time.Hour).SetVal("OK")

	err := pkg.WarmUpNotifyCache()
	assert.NoError(t, err)
	mockDB.ExpectationsWereMet(t)
	assert.NoError(t, rdbmock.ExpectationsWereMet())
}
