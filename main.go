package main

import (
	"fmt"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"log"
	"os"
	"saurfang/internal/config"
	"saurfang/internal/route"
)

var (
	BuildVersion   string
	BuildGoVersion string
	BuildTime      string
)

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
			config.InitMySQL()
			config.InitEtcd()
			if serve {
				app := fiber.New(fiber.Config{
					TrustProxy: true,
				})
				app.Use(logger.New())
				route.CMDBRouter(app)
				route.GameRouter(app)
				route.DataSourceRouter(app)
				route.TaskRouter(app)
				if err := app.Listen(fmt.Sprintf(":%s", os.Getenv("APP_PORT"))); err != nil {
					log.Fatalln("Failed to start app: ", err.Error())
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
