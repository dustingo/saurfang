package route

import (
	"github.com/gofiber/fiber/v3"
	"os"
	"saurfang/internal/config"
	"saurfang/internal/handler/taskhandler"
	"saurfang/internal/models/task.go"
	"saurfang/internal/models/upload"
	"saurfang/internal/repository/base"
)

type TaskRouteModule struct {
	Namespace string
	Comment   string
}

func (t *TaskRouteModule) Info() (namespace string, comment string) {
	namespace = t.Namespace
	comment = t.Comment
	return t.Namespace, t.Comment
}
func (t *TaskRouteModule) RegisterRoutesModule(r *fiber.App) {
	taskRouter := r.Group(t.Namespace)
	/*
		创建发布任务
	*/
	deployhandler := taskhandler.DeployHandler{base.BaseGormRepository[task.GameDeploymentTask]{DB: config.DB}}
	taskRouter.Post("/deploy/create", deployhandler.Handler_CreateDeployTask)
	taskRouter.Delete("/deploy/delete/:id", deployhandler.Handler_DeleteDeployTask)
	taskRouter.Get("/deploy/list", deployhandler.Handler_ShowDeployTask)
	taskRouter.Get("/deploy/listById/:id", deployhandler.Handler_ShowDeployTaskByID)
	taskRouter.Get("/deploy/listPerPage", deployhandler.Handler_ShowDeployPerPage)

	/*
		创建playbook
	*/
	playbookhandler := taskhandler.PlaybookHandler{base.NomadJobRepository{Consul: config.ConsulCli, Ns: os.Getenv("GAME_PLAYBOOK_NAMESPACE")}}
	taskRouter.Post("/playbook/create", playbookhandler.Handler_CreatePlaybook)
	taskRouter.Delete("/playbook/delete/:key", playbookhandler.Handler_DeletePlaybook)
	taskRouter.Put("/playbook/update", playbookhandler.Handler_UpdatePlaybook)
	taskRouter.Get("/playbook/list", playbookhandler.Handler_ShowPlaybook)
	taskRouter.Get("/playbook/listByKey", playbookhandler.Handler_ShowPlaybookByKey)
	taskRouter.Get("/playbook/select", playbookhandler.Handler_PlaybookSelect)

	/*
		上传服务器端
	*/
	uploadhandler := taskhandler.UploadHandler{base.BaseGormRepository[upload.UploadRecord]{DB: config.DB}}
	taskRouter.Get("/upload/file/list", uploadhandler.Handler_ShowServerPackage)
	taskRouter.Get("/upload/records", uploadhandler.Handler_ShowUploadRecords)
	taskRouter.Get("/upload/server", uploadhandler.Handler_UploadServerPackage)
	/*
		创建游戏服配置发布任务
	*/
	gameconfigDeployTaskHandler := taskhandler.ConfigDeployHandler{base.BaseGormRepository[task.ConfigDeployTask]{DB: config.DB}}
	taskRouter.Post("/config/create", gameconfigDeployTaskHandler.Handler_CreateConfigDeployTask)
	taskRouter.Delete("/config/delete/:id", gameconfigDeployTaskHandler.Handler_DeleteConfigDeployTask)
	taskRouter.Get("/config/list", gameconfigDeployTaskHandler.Handler_ShowConfigDeployTask)
	taskRouter.Get("/config/listPerPage", gameconfigDeployTaskHandler.Handler_ShowConfigDeployTaskPerPage)
	/*
		创建计划任务
	*/
	cronjobTaskHandler := taskhandler.CronjobHandler{base.BaseGormRepository[task.CronJobs]{DB: config.DB}}
	taskRouter.Post("/cronjob/create", cronjobTaskHandler.Handler_CreateCronjobTask)
	taskRouter.Delete("/cronjob/delete/:id", cronjobTaskHandler.Handler_DeleteCronjobTask)
	taskRouter.Put("/cronjob/update/:id", cronjobTaskHandler.Handler_UpdateCronjobTask)
	taskRouter.Get("/cronjob/list", cronjobTaskHandler.Handler_ShowCronjobTask)
	taskRouter.Put("/cronjob/reset/:id", cronjobTaskHandler.Handler_ResetCronjobStatus)
	/*
		执行任务
	*/
	// 执行游戏进程发布
	//taskRouter.Get("/run/process/deploy", deployhandler.Handler_RunGameDeployTask)
	//// 执行游戏配置发布
	//taskRouter.Get("/run/config/deploy", deployhandler.Handler_RunConfigDeployTask)
	//taskRouter.Get("/run/task/:id", deployhandler.Handler_RunOpsTask)
	/*
		常规任务
	*/
	normakTaskHandler := taskhandler.OpsTaskHandler{base.BaseGormRepository[task.SaurfangOpstask]{DB: config.DB}}
	taskRouter.Post("/ops/create", normakTaskHandler.Handler_CreateOpsNormalTask)
	taskRouter.Delete("/ops/delete/:id", normakTaskHandler.Handler_DeleteOpsNormalTask)
	taskRouter.Get("/ops/listPerPage", normakTaskHandler.Handler_ShowOpsNormalTaskPerPage)
	taskRouter.Get("/ops/select", normakTaskHandler.Handler_CrontabJobTaskSelect)
}
func init() {
	RegisterRoutesModule(&TaskRouteModule{Namespace: "/api/v1/task", Comment: "运维操作管理"})
}
