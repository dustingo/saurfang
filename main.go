package main

import (
	_ "embed"
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
	"log"
	"os"
	"os/signal"
	"saurfang/internal/config"
	"saurfang/internal/middleware"
	"saurfang/internal/route"
	"saurfang/internal/tools"
	"saurfang/internal/tools/pkg"
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
			if serve {
				app := fiber.New(fiber.Config{
					TrustProxy: true,
				})
				app.Use(cors.New())
				app.Use(recoverer.New())
				app.Use(limiter.New(limiter.Config{
					Max: 10,
					LimitReached: func(ctx fiber.Ctx) error {
						ctx.Set("Content-Type", "text/html; charset=utf-8")
						return ctx.Send([]byte(toofast))
					},
				}))
				app.Use(requestid.New())
				app.Use(logger.New())
				app.Use(middleware.UserAuth())
				for id, module := range route.RoutesModules {
					module.RegisterRoutesModule(app)
					namespace, comment := module.Info()
					modinfo := tools.PermissionData{
						ID:    uint(id),
						Name:  namespace,
						Group: comment,
					}
					tools.InitPermissionsItems(modinfo)
				}

				//app.Get("/debug", func(ctx fiber.Ctx) error {
				//	id := ctx.Query("userid")
				//	user_id, _ := strconv.Atoi(id)
				//	var res user.UserRole
				//	// 先确认用户归属组role
				//	if err := config.DB.Debug().Table("user_roles").Where("user_id = ?", uint(user_id)).First(&res).Error; err != nil {
				//		return ctx.Status(501).JSON(&fiber.Map{
				//			"code": 501,
				//			"msg":  err.Error,
				//		})
				//	}
				//	fmt.Println("roleid = ", res.RoleID)
				//	var perms []user.Permission
				//	if err := config.DB.Debug().Table("permissions p").Select("p.name", "p.group", "p.id").Joins("JOIN role_permissions rp ON rp.permission_id = p.id").
				//		Where("rp.role_id = ?", res.RoleID).Scan(&perms).Error; err != nil {
				//		return ctx.Status(501).JSON(&fiber.Map{
				//			"code": 501,
				//			"msg":  err.Error(),
				//		})
				//	}
				//	//if err := config.DB.Debug().Exec("select p.name,p.group,p.`id`  from permissions p JOIN `role_permissions` rp ON rp.`permission_id`  = p.id WHERE rp.`role_id`  = ? ", res.RoleID).Scan(&perms).Error; err != nil {
				//	//	return ctx.Status(501).JSON(&fiber.Map{
				//	//		"code": 501,
				//	//		"msg":  err.Error(),
				//	//	})
				//	//}
				//	//if err := config.DB.Debug().Table("role_permissions").Select("permissions,id,permissions.name,permissions.group").Joins("JOIN permissions ON role_permissions.permission_id = permissions.id").Where("role_permissions.role_id = ?", res.RoleID).Scan(&perms).Error; err != nil {
				//	//	return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				//	//		"code": fiber.StatusInternalServerError,
				//	//		"msg":  err.Error(),
				//	//	})
				//	//}
				//	return ctx.Status(200).JSON(fiber.Map{
				//		"code": 200,
				//		"data": perms,
				//	})
				//})
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
