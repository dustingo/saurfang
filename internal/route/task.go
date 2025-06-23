package route

import (
	"github.com/gofiber/fiber/v3"
	"os"
	"saurfang/internal/config"
	"saurfang/internal/handler/taskhandler"

	"saurfang/internal/service/taskservice"
)

func TaskRouter(app *fiber.App) *fiber.Router {
	taskRouter := app.Group("/api/v1/task")
	/*
		发布任务
	*/
	deployservice := taskservice.NewDeployService(config.DB)
	deployhandler := taskhandler.NewDeployHandler(deployservice)
	taskRouter.Post("/deploy/create/:id", deployhandler.Handler_CreateDeployTask)
	taskRouter.Delete("/deploy/delete/:id", deployhandler.Handler_DeleteDeployTask)
	taskRouter.Get("/deploy/list", deployhandler.Handler_ShowDeployTask)
	taskRouter.Get("/deploy/listById/:id", deployhandler.Handler_ShowDeployTaskByID)
	taskRouter.Get("/deploy/listPerPage", deployhandler.Handler_ShowDeployPerPage)

	/*
		playbook
	*/
	playbookservice := taskservice.NewPlaybookService(config.Etcd, os.Getenv("GAME_PLAYBOOK_NAMESPACE"))
	playbookhandler := taskhandler.NewPlaybookHandler(playbookservice)
	taskRouter.Post("/playbook/create", playbookhandler.Handler_CreatePlaybook)
	taskRouter.Delete("/playbook/delete/:key", playbookhandler.Handler_DeletePlaybook)
	taskRouter.Put("/playbook/update", playbookhandler.Handler_UpdatePlaybook)
	taskRouter.Get("/playbook/list", playbookhandler.Handler_ShowPlaybook)
	taskRouter.Get("/playbook/listByKey", playbookhandler.Handler_ShowPlaybookByKey)
	taskRouter.Get("/playbook/select", playbookhandler.Handler_PlaybookSelect)

	/*
		上传服务器端
	*/
	uploadservice := taskservice.NewUploadService(config.DB)
	uploadhandler := taskhandler.NewUploadHandler(uploadservice)
	taskRouter.Get("/upload/file/list", uploadhandler.Handler_ShowServerPackage)
	taskRouter.Get("/upload/records", uploadhandler.Handler_ShowUploadRecords)
	taskRouter.Get("/upload/server", uploadhandler.Handler_UploadServerPackage)
	/*
		游戏服配置发布任务
	*/
	gameconfigDeployTaskService := taskservice.NewConfigDeployService(config.DB)
	gameconfigDeployTaskHandler := taskhandler.NewConfigDeployHandler(gameconfigDeployTaskService)
	taskRouter.Post("/config/create", gameconfigDeployTaskHandler.Handler_CreateConfigDeployTask)
	taskRouter.Delete("/config/delete/:id", gameconfigDeployTaskHandler.Handler_DeleteConfigDeployTask)
	taskRouter.Get("/config/list", gameconfigDeployTaskHandler.Handler_ShowConfigDeployTask)

	/*
		执行任务
	*/
	// 执行游戏进程发布
	taskRouter.Get("/run/process/deploy", deployhandler.Handler_RunGameDeployTask)
	taskRouter.Get("/run/config/deploy", deployhandler.Handler_RunConfigDeployTask)
	taskRouter.Get("/run/task/:id", deployhandler.Handler_RunOpsTask)
	/*
		常规任务
	*/
	normalTaskService := taskservice.NewOpsTaskService(config.DB)
	normakTaskHandler := taskhandler.NewOpsTaskHandler(normalTaskService)
	taskRouter.Post("/ops/create", normakTaskHandler.Handler_CreateOpsNormalTask)
	taskRouter.Delete("/ops/delete/:id", normakTaskHandler.Handler_DeleteOpsNormalTask)
	taskRouter.Get("/ops/listPerPage", normakTaskHandler.Handler_ShowOpsNormalTaskPerPage)
	taskRouter.Get("/ops/select", normakTaskHandler.Handler_CrontabJobTaskSelect)
	return &taskRouter
}
