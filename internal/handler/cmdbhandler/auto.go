package cmdbhandler

import (
	"fmt"
	"saurfang/internal/config"
	"saurfang/internal/models/amis"
	"saurfang/internal/models/autosync"
	"saurfang/internal/models/gamehost"
	"saurfang/internal/repository/base"
	"saurfang/internal/tools/pkg"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

type AutoSyncHandler struct {
	base.BaseGormRepository[autosync.AutoSync]
}

// Handler_CreateAutoSyncConfig handler_CreateAutoSyncConfig 创建自动同步配置
// tested
func (a *AutoSyncHandler) Handler_CreateAutoSyncConfig(c fiber.Ctx) error {
	var config autosync.AutoSync
	if err := c.Bind().Body(&config); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "failed to create auto sync config", err.Error(), fiber.Map{})
	}
	if err := a.Create(&config); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to create auto sync config", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", fiber.Map{})
}

// Handler_ShowAutoSyncConfig handler_ShowAutoSyncConfig 展示自动同步配置
// tested
func (a *AutoSyncHandler) Handler_ShowAutoSyncConfig(c fiber.Ctx) error {
	configs, err := a.List()
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to show auto sync config", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", fiber.Map{
		"data": configs,
	})
}

// Handler_UpdateAutoSyncConfig handler_UpdateAutoSyncConfig 更新自动同步配置
// tested
func (a *AutoSyncHandler) Handler_UpdateAutoSyncConfig(c fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var syncConfig autosync.AutoSync
	if err := c.Bind().Body(&syncConfig); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "failed to update auto sync config", err.Error(), fiber.Map{})
	}
	if uint(id) != syncConfig.ID {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "failed to update auto sync config", "id is the same as the sync config", fiber.Map{})
	}
	if err := a.UpdateALL(&syncConfig); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to update auto sync config", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", fiber.Map{})
}

// Handler_DeleteAutoSyncConfig handler_DeleteAutoSyncConfig 删除自动同步配置
// tested
func (a *AutoSyncHandler) Handler_DeleteAutoSyncConfig(c fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := a.Delete(uint(id)); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to delete auto sync config", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", nil)
}

// Handler_AutoSync handler_AutoSync 自动同步
func (a *AutoSyncHandler) Handler_AutoSync(c fiber.Ctx) error {
	var target = struct {
		Target string `json:"target"`
	}{}
	if err := c.Bind().Body(&target); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "failed to auto sync", err.Error(), fiber.Map{})
	}
	targetLabel := strings.Split(target.Target, "-")
	if len(targetLabel) < 2 {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "failed to auto sync", target.Target+" is invalid", fiber.Map{})
	}
	if targetLabel[0] == "阿里云" {
		if err := a.AutoSyncAliYunEcs(targetLabel[1]); err != nil {
			return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to auto sync", err.Error(), fiber.Map{})
		}
	} else {
		if err := a.AutoSyncHuaweiECS(targetLabel[1]); err != nil {
			return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to auto sync", err.Error(), fiber.Map{})
		}
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", nil)
}

// Handler_AutoSyncConfigSelect handler_AutoSyncConfigSelect 自动同步配置选择
// tested
func (a *AutoSyncHandler) Handler_AutoSyncConfigSelect(c fiber.Ctx) error {
	configs, err := a.List()
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to auto sync config select", err.Error(), fiber.Map{})
	}
	var op amis.AmisOptionsString
	var ops []amis.AmisOptionsString
	for _, sn := range configs {
		op.Label = sn.Label
		op.SelectMode = "tree"
		op.Value = sn.Cloud + "-" + sn.Label
		ops = append(ops, op)
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", fiber.Map{
		"options": ops,
	})
}
func (a *AutoSyncHandler) AutoSyncAliYunEcs(target string) error {
	var autoConfig autosync.AutoSync
	if err := a.DB.First(&autoConfig, "label = ?", target).Error; err != nil {
		return err
	}
	client, err := pkg.NewAlyumClient(autoConfig.AccessKey, autoConfig.SecretKey, autoConfig.Endpoint, autoConfig.Region, autoConfig.GroupID)
	if err != nil {
		return err
	}
	cloudServers, err := client.Fetch()
	if err != nil {
		return err
	}
	// 获取本地ecs数据
	// 解析云数据，将数据映射都本地数据结构
	// 对比本地与云数据,新增及修改
	var localServers []gamehost.Hosts
	if err := config.DB.Raw("select * from hosts").Scan(&localServers).Error; err != nil {
		return err
	}
	//
	localServerMap := make(map[string]string)
	//remoteServerMap := make(map[string]string)
	for _, localserver := range localServers {
		// instance,cpu,mem,inip,outerip,name
		hashData := pkg.Hash(fmt.Sprintf("%s-%s-%s-%s-%s-%s", localserver.InstanceID, localserver.CPU, localserver.Memory, localserver.PrivateIP, localserver.PublicIP, localserver.Hostname))
		localServerMap[localserver.InstanceID] = hashData

	}
	sqls := []string{}
	for _, cloudServer := range cloudServers {
		for _, server := range cloudServer.Instances.Instance {
			if hashstring, ok := localServerMap[server.InstanceId]; ok {
				if pkg.Hash(fmt.Sprintf("%s-%s-%s-%s-%s-%s", server.InstanceId, strconv.Itoa(server.Cpu), strconv.Itoa(server.Memory), pkg.SingleIP(server.VpcAttributes.PrivateIpAddress.IPAddress), pkg.SingleIP(server.PublicIpAddress.IpAddress), server.InstanceName)) != hashstring {
					sql := fmt.Sprintf("UPDATE hosts SET hostname = '%s',public_ip = '%s',private_ip = '%s',cpu='%s',memory='%s',os_name='%s' WHERE instance_id = '%s';",
						server.InstanceName, pkg.SingleIP(server.PublicIpAddress.IpAddress), pkg.SingleIP(server.VpcAttributes.PrivateIpAddress.IPAddress), strconv.Itoa(server.Cpu),
						strconv.Itoa(server.Memory), server.OSName, server.InstanceId)
					sqls = append(sqls, sql)
				}
				// 开始映射到本地结构体，插入或修改 update
			} else {
				sql := fmt.Sprintf("INSERT INTO hosts (created_at, instance_id, hostname, public_ip, private_ip, cpu, memory, os_name,instance_type,port) VALUES('%v','%s','%s','%s','%s','%s','%s','%s','%s','22');", time.Now().Format("2006-01-02T15:04:05"),

					server.InstanceId, server.InstanceName, pkg.SingleIP(server.PublicIpAddress.IpAddress), pkg.SingleIP(server.VpcAttributes.PrivateIpAddress.IPAddress), strconv.Itoa(server.Cpu),

					strconv.Itoa(server.Memory), server.OSName, server.InstanceType)
				sqls = append(sqls, sql)
			}
		}
	}
	orm := config.DB
	tx := orm.Begin()
	for _, sql := range sqls {
		if err := tx.Exec(sql).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

func (a *AutoSyncHandler) AutoSyncHuaweiECS(target string) error {
	var autoConfig autosync.AutoSync
	if err := a.DB.First(&autoConfig, "label = ?", target).Error; err != nil {
		return err
	}
	client, err := pkg.NewHuaweiClient(autoConfig.AccessKey, autoConfig.SecretKey, autoConfig.Region, autoConfig.GroupID)
	if err != nil {
		return err
	}
	cloudServers, err := client.Fetch()
	if err != nil {
		return err
	}
	var localServers []gamehost.Hosts
	if err := config.DB.Raw("select * from hosts").Scan(&localServers).Error; err != nil {
		return err
	}
	localServerMap := make(map[string]string)
	//remoteServerMap := make(map[string]string)
	for _, localserver := range localServers {
		// instance,cpu,mem,inip,outerip,name
		hashData := pkg.Hash(fmt.Sprintf("%s-%s-%s-%s-%s-%s", localserver.InstanceID, localserver.CPU, localserver.Memory, localserver.PrivateIP, localserver.PublicIP, localserver.Hostname))
		localServerMap[localserver.InstanceID] = hashData

	}
	sqls := []string{}
	var private_ip, public_ip string
	for _, cloudServer := range *cloudServers {
		for _, server := range cloudServer.Servers {
			for _, ips := range server.Addresses[server.MetaData.VpcID] {
				switch ips.OSEXTIPStype {
				case "fixed":
					private_ip = ips.Addr
				default:
					public_ip = ips.Addr
				}
			}
			if hashstring, ok := localServerMap[server.Id]; ok {
				if pkg.Hash(fmt.Sprintf("%s-%s-%s-%s-%s-%s", server.Id, strconv.Itoa(server.Flavor.Vcpus), strconv.Itoa(server.Flavor.Ram), private_ip, public_ip, server.Name)) != hashstring {
					sql := fmt.Sprintf("UPDATE hosts SET hostname = '%s',public_ip = '%s',private_ip = '%s',cpu='%s',memory='%s',os_name='%s' WHERE instance_id = '%s';",
						server.Name, public_ip, private_ip, strconv.Itoa(server.Flavor.Vcpus),
						strconv.Itoa(server.Flavor.Ram), server.MetaData.ImageName, server.Id)
					sqls = append(sqls, sql)
				}
				// 开始映射到本地结构体，插入或修改 update
			} else {
				sql := fmt.Sprintf("INSERT INTO hosts (created_at, instance_id, hostname, public_ip, private_ip, cpu, memory, os_name,instance_type,port) VALUES('%v','%s','%s','%s','%s','%s','%s','%s','%s','22');", time.Now().Format("2006-01-02T15:04:05"),
					server.Id, server.Name, public_ip, private_ip, strconv.Itoa(server.Flavor.Vcpus),
					strconv.Itoa(server.Flavor.Ram), server.Name, server.Flavor.Name)
				sqls = append(sqls, sql)
			}
		}
	}
	orm := config.DB
	tx := orm.Begin()
	for _, sql := range sqls {
		if err := tx.Exec(sql).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}
