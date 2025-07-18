package main

import (
	_ "embed"
	"errors"
	"fmt"
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
	"log"
	"os"
	"os/signal"
	"saurfang/internal/config"
	"saurfang/internal/middleware"
	"saurfang/internal/models/autosync"
	"saurfang/internal/models/credential"
	"saurfang/internal/models/dashboard"
	"saurfang/internal/models/datasource"
	"saurfang/internal/models/gamechannel"
	"saurfang/internal/models/gamegroup"
	"saurfang/internal/models/gamehost"
	"saurfang/internal/models/gameserver"
	"saurfang/internal/models/task.go"
	"saurfang/internal/models/upload"
	"saurfang/internal/models/user"
	"saurfang/internal/route"
	"saurfang/internal/tools"
	"saurfang/internal/tools/pkg"
	"strings"
	"syscall"
)

var (
	BuildVersion   string
	BuildGoVersion string
	BuildTime      string
)

//go:embed web/toofast.html
var toofast string

func main() {
	var (
		serve, migrate, generate bool
		char                     string
	)
	var rootCmd = &cobra.Command{
		Use:          "--serve| --migrate | -- generate --char| --release",
		SilenceUsage: false,
		PreRun: func(cmd *cobra.Command, args []string) {
			fmt.Printf("go version: %s\nBuild version: %s\nBuild time: %s\n",
				BuildGoVersion, BuildVersion, BuildTime)
		},
		Run: func(cmd *cobra.Command, args []string) {
			err := godotenv.Load(".env")
			if err != nil {
				log.Fatalln("无法加载.env文件:", err)
			}
			// 目录初始化
			if err := tools.PathInit([]string{os.Getenv("SERVER_PACKAGE_SRC_PATH"), os.Getenv("SERVER_PACKAGE_DEST_PATH")}); err != nil {
				log.Fatalln("Init uploadpath dir failed:", err.Error())
			}
			config.InitMySQL()
			config.InitEtcd()
			config.InitSynq()
			config.InitCache()
			if serve {
				app := fiber.New(fiber.Config{
					TrustProxy:  true,
					ProxyHeader: fiber.HeaderXForwardedFor,
					TrustProxyConfig: fiber.TrustProxyConfig{
						Proxies: strings.Split(os.Getenv("APP_TRUST_PROXY"), ","),
					},
				})
				app.Use(cors.New())
				app.Use(recoverer.New())
				app.Use(limiter.New(limiter.Config{
					Max: 200,
					LimitReached: func(ctx fiber.Ctx) error {
						ctx.Set("Content-Type", "text/html; charset=utf-8")
						return ctx.Send([]byte(toofast))
					},
				}))
				app.Use(requestid.New())
				app.Use(middleware.UserAuth())
				app.Use(logger.New(logger.Config{
					CustomTags: map[string]logger.LogFunc{
						"requestid": func(output logger.Buffer, c fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
							return output.WriteString(requestid.FromContext(c))
						},
						"user": func(output logger.Buffer, c fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
							return output.WriteString(c.Get("X-Request-User"))
						},
					},
					Format: "${time} ${user} ${requestid} ${ip} ${status} - ${latency} ${method} ${path} ${error}\n",
				}))

				for _, module := range route.RoutesModules {
					module.RegisterRoutesModule(app)
					namespace, comment := module.Info()
					modinfo := tools.PermissionData{
						Name:  namespace,
						Group: comment,
					}
					tools.InitPermissionsItems(&modinfo)
					fmt.Println("路由组: ", namespace, "别名: ", comment)

				}
				// 加载权限到缓存
				pkg.WarmUpCache()
				go pkg.TaskManagerSetup()
				go pkg.CheckActiveInterval(config.DB)
				go func() {
					if err := app.Listen(fmt.Sprintf(":%s", os.Getenv("APP_PORT"))); err != nil {
						log.Fatalln("Failed to start app: ", err.Error())
					}
				}()
				var queues = make(map[string]int)
				queues[config.SynqConfig.Queue] = 6
				synqSrv := asynq.NewServer(
					asynq.RedisClientOpt{Addr: config.SynqConfig.Addr, Password: config.SynqConfig.Password, DB: config.SynqConfig.DB},
					asynq.Config{
						Concurrency: 10,
						Queues:      queues,
					},
				)
				mux := asynq.NewServeMux()
				mux.HandleFunc("ops", pkg.CronTaskHandler)
				go func() {
					if err := synqSrv.Run(mux); err != nil {
						log.Fatalln("Failed to start synq:", err.Error())
					}
				}()
				quit := make(chan os.Signal, 1)
				signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
				<-quit
				if err := app.Shutdown(); err != nil {
					log.Fatalln(fmt.Errorf("shutdown app failed: %v", err))
				}
				synqSrv.Shutdown()
			}
			if migrate {
				if err := config.DB.AutoMigrate(&credential.UserCredential{}, &upload.UploadRecord{}, &user.User{}, &user.Role{},
					&gamehost.SaurfangHosts{}, &gamechannel.SaurfangChannels{}, &gamegroup.SaurfangGroups{}, &gameserver.SaurfangGames{},
					&gameserver.SaurfangGameHosts{}, &datasource.SaurfangDatasources{}, &task.CronJobs{}, &task.SaurfangPublishtask{},
					&task.SaurfangGameconfigtask{}, &dashboard.SaurfangTaskdashboards{}, &dashboard.LoginRecords{}, &dashboard.ResourceStatistics{},
					&autosync.SaurfangAutoSync{}, &user.InviteCodes{}); err != nil {
					log.Fatalln("AutoMigrate failed:", err)
				}
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
							if err := config.DB.Table("roles").Create(&role).Error; err != nil {
								log.Fatalln("Create role failed:", err)
							}
						} else {
							log.Fatalln("failed to query role:", err)
						}
					}
				}
			}
			if generate {
				if err := tools.GenerateInviteCodes(char); err != nil {
					log.Fatalln("GenerateInviteCodes failed:", err)
				}
			}
		},
	}
	rootCmd.Flags().BoolVar(&serve, "serve", false, "启动服务器")
	rootCmd.Flags().BoolVar(&migrate, "migrate", false, "初始化数据库")
	rootCmd.Flags().BoolVar(&generate, "generate", false, "生成注册邀请码")
	rootCmd.Flags().StringVar(&char, "char", "", "生成注册邀请码的基本字符串")
	rootCmd.MarkFlagsMutuallyExclusive("serve", "migrate", "generate")
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
