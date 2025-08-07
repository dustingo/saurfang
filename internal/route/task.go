package route

import (
	"saurfang/internal/config"
	"saurfang/internal/handler/taskhandler"
	"saurfang/internal/models/task"
	"saurfang/internal/models/upload"
	"saurfang/internal/repository/base"

	"github.com/gofiber/fiber/v3"
)

type TaskRouteModule struct {
	Namespace string
	Comment   string
}

func (t *TaskRouteModule) Info() (namespace string, comment string) {
	// namespace = t.Namespace
	// comment = t.Comment
	return t.Namespace, t.Comment
}
func (t *TaskRouteModule) RegisterRoutesModule(r *fiber.App) {
	taskRouter := r.Group(t.Namespace)
	/*
		创建发布任务
	*/
	deployhandler := taskhandler.DeployHandler{BaseGormRepository: base.BaseGormRepository[task.GameDeploymentTask]{DB: config.DB}}
	taskRouter.Post("/deploy/create", deployhandler.Handler_CreateDeployTask)
	taskRouter.Delete("/deploy/delete/:id", deployhandler.Handler_DeleteDeployTask)
	taskRouter.Get("/deploy/list", deployhandler.Handler_ShowDeployTask)
	taskRouter.Get("/deploy/listById/:id", deployhandler.Handler_ShowDeployTaskByID)
	taskRouter.Get("/deploy/listPerPage", deployhandler.Handler_ShowDeployPerPage)

	/*
		上传服务器端
	*/
	uploadhandler := taskhandler.UploadHandler{BaseGormRepository: base.BaseGormRepository[upload.UploadRecord]{DB: config.DB}}
	taskRouter.Get("/upload/file/list", uploadhandler.Handler_ShowServerPackage)
	taskRouter.Get("/upload/records", uploadhandler.Handler_ShowUploadRecords)
	taskRouter.Get("/upload/server", uploadhandler.Handler_UploadServerPackage)

	/*
		创建计划任务
	*/
	cronjobTaskHandler := taskhandler.CronjobHandler{BaseGormRepository: base.BaseGormRepository[task.CronJobs]{DB: config.DB}}
	taskRouter.Post("/cronjob/create", cronjobTaskHandler.Handler_CreateCronjobTask)
	taskRouter.Delete("/cronjob/delete/:id", cronjobTaskHandler.Handler_DeleteCronjobTask)
	taskRouter.Put("/cronjob/update/:id", cronjobTaskHandler.Handler_UpdateCronjobTask)
	taskRouter.Get("/cronjob/list", cronjobTaskHandler.Handler_ShowCronjobTask)
	taskRouter.Put("/cronjob/reset/:id", cronjobTaskHandler.Handler_ResetCronjobStatus)

	// 新增：获取可用的自定义任务和游戏服务器列表
	taskRouter.Get("/cronjob/available-custom-tasks", cronjobTaskHandler.Handler_GetAvailableCustomTasks)
	taskRouter.Get("/cronjob/available-servers", cronjobTaskHandler.Handler_GetAvailableServers)

	/*
		自定义任务管理
	*/
	customTaskHandler := taskhandler.NewCustomTaskHandler()
	taskRouter.Post("/custom/create", customTaskHandler.Handler_CreateCustomTask)
	taskRouter.Get("/custom/list", customTaskHandler.Handler_ListCustomTasks)

	// 执行状态管理
	taskRouter.Get("/custom/monitor/execution/status/:execution_id", customTaskHandler.Handler_GetExecutionStatus)
	taskRouter.Get("/custom/monitor/execution/logs/:execution_id", customTaskHandler.Handler_GetExecutionLogs)
	taskRouter.Post("/custom/monitor/execution/stop/:execution_id", customTaskHandler.Handler_StopExecution)
	taskRouter.Get("/custom/monitor/executions/:task_id", customTaskHandler.Handler_ListExecutions)

	//查询自定义任务执行记录
	customTaskExecutionsHandler := taskhandler.NewCustomTaskExecutionsHandler()
	taskRouter.Get("/custom/record/executions", customTaskExecutionsHandler.Handler_ListCustomExecutions)
	taskRouter.Delete("/custom/record/executions/:id", customTaskExecutionsHandler.Handler_DeleteCustomExecutions)

	// 这些路由必须放在更具体的路由之后，避免冲突
	taskRouter.Get("/custom/listById/:id", customTaskHandler.Handler_GetCustomTask)
	taskRouter.Put("/custom/update/:id", customTaskHandler.Handler_UpdateCustomTask)
	taskRouter.Delete("/custom/delete/:id", customTaskHandler.Handler_DeleteCustomTask)
	// 执行自定义任务
	taskRouter.Post("/custom/execute/:id", customTaskHandler.Handler_ExecuteCustomTask)
}
func init() {
	RegisterRoutesModule(&TaskRouteModule{Namespace: "/api/v1/task", Comment: "运维操作管理"})
}
