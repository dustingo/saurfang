package main

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"saurfang/internal/config"
	"saurfang/internal/handler/taskhandler"
	"saurfang/internal/middleware"
	"saurfang/internal/models/autosync"
	"saurfang/internal/models/credential"
	"saurfang/internal/models/dashboard"
	"saurfang/internal/models/datasource"
	"saurfang/internal/models/gamechannel"
	"saurfang/internal/models/gamegroup"
	"saurfang/internal/models/gamehost"
	"saurfang/internal/models/gameserver"
	"saurfang/internal/models/notify"
	"saurfang/internal/models/task"
	"saurfang/internal/models/upload"
	"saurfang/internal/models/user"
	"saurfang/internal/route"
	"saurfang/internal/tools"
	"saurfang/internal/tools/ntfy"
	"saurfang/internal/tools/pkg"
	"strings"
	"syscall"

	_ "go.uber.org/automaxprocs"
	"golang.org/x/crypto/bcrypt"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/logger"
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var (
	BuildVersion   string
	BuildGoVersion string
	BuildTime      string
)

//go:embed web/toofast.html
var toofast string
var (
	serve, migrate, generate bool
	char                     string
	setAdmin                 string
)

func main() {
	rootCmd := createRootCommand()
	adminCmd := createAdminCommand()

	// 配置根命令标志
	rootCmd.Flags().BoolVar(&serve, "serve", false, "启动服务器")
	rootCmd.Flags().BoolVar(&migrate, "migrate", false, "初始化数据库")
	rootCmd.Flags().BoolVar(&generate, "generate", false, "生成注册邀请码")
	rootCmd.Flags().StringVar(&char, "char", "", "生成注册邀请码的基本字符串")
	rootCmd.MarkFlagsMutuallyExclusive("serve", "migrate", "generate")

	// 配置管理员命令标志
	adminCmd.Flags().StringVar(&setAdmin, "set-admin", "", "设置管理员")
	adminCmd.MarkFlagRequired("set-admin")

	// 添加子命令
	rootCmd.AddCommand(adminCmd)

	// 执行命令
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

// createRootCommand 创建根命令
func createRootCommand() *cobra.Command {
	return &cobra.Command{
		Use:          "saurfang",
		Short:        "Saurfang游戏运维工具",
		Long:         "Saurfang是一个用于游戏服务器运维管理的工具，支持服务器启动、数据库迁移、邀请码生成等功能。",
		SilenceUsage: true,
		PreRun:       printVersionInfo,
		Run:          runRootCommand,
	}
}

// createAdminCommand 创建管理员命令
func createAdminCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "admin",
		Short: "管理员相关操作",
		Long:  "管理员相关操作，包括设置管理员用户等功能。",
		Run:   runAdminCommand,
	}
}

// printVersionInfo 打印版本信息
func printVersionInfo(cmd *cobra.Command, args []string) {
	fmt.Printf("Go版本: %s\n构建版本: %s\n构建时间: %s\n",
		BuildGoVersion, BuildVersion, BuildTime)
}

// runRootCommand 执行根命令
func runRootCommand(cmd *cobra.Command, args []string) {
	// 启动pprof服务
	//startPprofServer()

	// 加载环境变量
	loadEnvFile()

	// 初始化基础配置
	initializeBasicConfig()

	// 根据标志执行相应操作
	if serve {
		startServer()
	}
	if migrate {
		runDatabaseMigration()
	}
	if generate {
		generateInviteCodes()
	}
}

// runAdminCommand 执行管理员命令
func runAdminCommand(cmd *cobra.Command, args []string) {
	// 加载环境变量
	loadEnvFile()

	// 初始化数据库和缓存
	config.InitMySQL()
	config.InitCache()

	// 设置管理员
	if setAdmin != "" {
		setAdminFunc()
	}
}

// startPprofServer 启动pprof性能分析服务
func startPprofServer() {
	go func() {
		slog.Info("pprof server started on port 30552", "error", http.ListenAndServe(":30552", nil))
	}()
}

// loadEnvFile 加载环境变量文件
func loadEnvFile() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalln("无法加载.env文件:", err)
	}
}

// initializeBasicConfig 初始化基础配置
func initializeBasicConfig() {
	// 目录初始化
	if err := tools.PathInit([]string{os.Getenv("SERVER_PACKAGE_SRC_PATH"), os.Getenv("SERVER_PACKAGE_DEST_PATH")}); err != nil {
		log.Fatalln("Init uploadpath dir failed:", err.Error())
	}

	// 初始化各种配置
	config.InitMySQL()
	config.InitSynq()
	config.InitCache()
	config.InitNtfy()
	config.InitConsul()
}

// startServer 启动服务器
func startServer() {
	// 配置Fiber应用
	app := setupFiberApp()

	// 设置中间件
	setupMiddlewares(app)

	// 注册路由
	registerRoutes(app)

	// 初始化缓存和后台服务
	initializeServices()

	// 启动Web服务器
	startWebServer(app)

	// 启动异步任务服务器
	synqSrv := startAsynqServer()

	// 等待退出信号
	waitForShutdown(app, synqSrv)
}

// setupFiberApp 配置Fiber应用
func setupFiberApp() *fiber.App {
	// 获取信任代理配置
	trustProxyStr := os.Getenv("APP_TRUST_PROXY")
	var trustProxyConfig fiber.TrustProxyConfig

	if trustProxyStr != "" {
		// 如果设置了环境变量，使用指定的代理
		trustProxyConfig = fiber.TrustProxyConfig{
			Proxies: strings.Split(trustProxyStr, ","),
		}
	} else {
		// 如果没有设置，信任所有代理（适用于开发环境）
		trustProxyConfig = fiber.TrustProxyConfig{
			Proxies: []string{"0.0.0.0/0", "::/0"},
		}
	}

	return fiber.New(fiber.Config{
		TrustProxy:       true,
		ProxyHeader:      fiber.HeaderXForwardedFor,
		TrustProxyConfig: trustProxyConfig,
	})
}

// setupMiddlewares 设置中间件
func setupMiddlewares(app *fiber.App) {
	// 跨域
	app.Use(cors.New())

	// 恢复
	app.Use(recoverer.New())

	// 限流
	app.Use(limiter.New(limiter.Config{
		Max: 200,
		LimitReached: func(ctx fiber.Ctx) error {
			ctx.Set("Content-Type", "text/html; charset=utf-8")
			return ctx.Send([]byte(toofast))
		},
	}))

	// 请求ID
	app.Use(requestid.New())

	// 用户认证
	app.Use(middleware.UserAuth())

	// 日志
	app.Use(logger.New(logger.Config{
		CustomTags: map[string]logger.LogFunc{
			"requestid": func(output logger.Buffer, c fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
				return output.WriteString(requestid.FromContext(c))
			},
			"user": func(output logger.Buffer, c fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
				return output.WriteString(c.Get("X-Request-User"))
			},
			"query_string": func(output logger.Buffer, c fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
				// 只获取查询参数字符串
				queryString := string(c.Request().URI().QueryString())
				if queryString != "" {
					return output.WriteString("?" + queryString)
				}
				return 0, nil
			},
		},
		Format: "${time} ${user} ${requestid} ${ip} ${status} - ${latency} ${method} ${path}${query_string} ${error}\n",
	}))
}

// registerRoutes 注册路由
func registerRoutes(app *fiber.App) {
	for _, module := range route.RoutesModules {
		module.RegisterRoutesModule(app)
		namespace, comment := module.Info()
		modinfo := tools.PermissionData{
			Name:  namespace,
			Group: comment,
		}
		tools.InitPermissionsItems(&modinfo)
		fmt.Println("路由组:", namespace, "别名:", comment)
	}
}

// initializeServices 初始化服务
func initializeServices() {
	// 加载权限到缓存
	pkg.WarmUpCache()
	// 加载消息通知配置到缓存
	pkg.WarmUpNotifyCache()
	// 启动计划任务管理器
	go pkg.TaskManagerSetup()
	// 启动通知订阅监听器
	go ntfy.StartNotifySubscriber()
}

// startWebServer 启动Web服务器
func startWebServer(app *fiber.App) {
	go func() {
		if err := app.Listen(fmt.Sprintf(":%s", os.Getenv("APP_PORT"))); err != nil {
			log.Fatalln("Failed to start app:", err.Error())
		}
	}()
}

// startAsynqServer 启动异步任务服务器
func startAsynqServer() *asynq.Server {
	var queues = make(map[string]int)
	queues[config.SynqConfig.Queue] = 6

	synqSrv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     config.SynqConfig.Addr,
			Password: config.SynqConfig.Password,
			DB:       config.SynqConfig.DB,
		},
		asynq.Config{
			Concurrency: 10,
			Queues:      queues,
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc("custom_task", taskhandler.CustonCronjobHandler)
	mux.HandleFunc("server_op", pkg.ServerOperationHandler)

	go func() {
		if err := synqSrv.Run(mux); err != nil {
			log.Fatalln("Failed to start synq:", err.Error())
		}
	}()

	return synqSrv
}

// waitForShutdown 等待关闭信号并优雅关闭
func waitForShutdown(app *fiber.App, synqSrv *asynq.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务器...")

	if err := app.Shutdown(); err != nil {
		log.Fatalln(fmt.Errorf("shutdown app failed: %v", err))
	}

	synqSrv.Shutdown()
	log.Println("服务器已关闭")
}

// runDatabaseMigration 执行数据库迁移
func runDatabaseMigration() {
	// 执行数据库迁移
	if err := config.DB.AutoMigrate(
		&credential.UserCredential{}, &upload.UploadRecord{}, &user.User{}, &user.Role{},
		&gamehost.Hosts{}, &gamechannel.Channels{}, &gamegroup.Groups{}, &gameserver.Games{},
		&gameserver.GameHosts{}, &datasource.Datasources{}, &task.CronJobs{}, &task.GameDeploymentTask{},
		&dashboard.TaskDashboards{}, &dashboard.LoginRecords{}, &dashboard.ResourceStatistics{},
		&autosync.AutoSync{}, &user.InviteCodes{}, &task.CustomTask{}, &task.CustomTaskExecution{},
		notify.NotifySubscribe{}, notify.NotifyConfig{},
	); err != nil {
		log.Fatalln("AutoMigrate failed:", err)
	}

	// 创建默认角色
	createDefaultRoles()

	log.Println("数据库迁移完成")
}

// createDefaultRoles 创建默认角色
func createDefaultRoles() {
	defaultRoles := []user.Role{
		{Name: "管理员"},
		{Name: "运维"},
		{Name: "研发"},
		{Name: "未指定"},
	}

	for _, role := range defaultRoles {
		var existing user.Role
		if err := config.DB.Table("roles").Where("name = ?", role.Name).First(&existing).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err = config.DB.Table("roles").Create(&role).Error; err != nil {
					log.Fatalln("Create role failed:", err)
				}
				log.Printf("创建默认角色: %s", role.Name)
			} else {
				log.Fatalln("failed to query role:", err)
			}
		}
	}
}

// generateInviteCodes 生成邀请码
func generateInviteCodes() {
	if err := tools.GenerateInviteCodes(char); err != nil {
		log.Fatalln("GenerateInviteCodes failed:", err)
	}
	log.Println("邀请码生成完成")
}

// setAdminFunc 设置管理员用户
func setAdminFunc() {
	var existingUser user.User
	if err := config.DB.Table("users").Where("username = ?", setAdmin).First(&existingUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("用户 %s 不存在，无法设置为管理员", setAdmin)
			// 创建用户
			// 生成随机密码并使用bcrypt加密
			// 生成10位随机密码
			const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
			rawPassword := make([]byte, 10)
			for i := range rawPassword {
				rawPassword[i] = letters[rand.Intn(len(letters))]
			}
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
			if err != nil {
				log.Fatalln("failed to hash password:", err)
			}
			newUser := struct {
				ID       uint
				Username string
				Password string
			}{
				Username: setAdmin,
				Password: string(hashedPassword),
			}
			result := config.DB.Table("users").Where("username = ?", setAdmin).FirstOrCreate(&newUser)
			if result.Error != nil {
				log.Fatalln("failed to create user:", result.Error)
			}
			log.Printf("成功创建用户 %s,默认密码: %s", setAdmin, string(rawPassword))
			// 将用户设置为管理员
			if err := config.DB.Table("user_roles").Create(&user.UserRole{
				UserID: newUser.ID,
				RoleID: 1,
			}).Error; err != nil {
				log.Fatalln("failed to create user role:", err)
			}
			log.Printf("成功将用户 %s 设置为管理员", setAdmin)
			return
		}
		log.Fatalln("查询用户失败:", err)
	}
}
