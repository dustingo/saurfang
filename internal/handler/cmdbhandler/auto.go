package cmdbhandler

import (
	"fmt"
	"github.com/gofiber/fiber/v3"
	"saurfang/internal/config"
	"saurfang/internal/models/amis"
	"saurfang/internal/models/autosync"
	"saurfang/internal/models/gamehost"
	"saurfang/internal/service/cmdbservice"
	"saurfang/internal/tools/pkg"
	"strconv"
	"strings"
	"time"
)

type AutoSyncHandler struct {
	cmdbservice.AutoSyncService
}

func NewAutoSyncHandler(svc *cmdbservice.AutoSyncService) *AutoSyncHandler {
	return &AutoSyncHandler{*svc}
}

// handler_CreateAutoSyncConfig 创建自动同步配置
func (a *AutoSyncHandler) Handler_CreateAutoSyncConfig(c fiber.Ctx) error {
	var config autosync.SaurfangAutoSync
	if err := c.Bind().Body(&config); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": 1,
			"msg":    err.Error(),
		})
	}
	if err := a.Create(&config); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": 1,
			"msg":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
	})
}

func (a *AutoSyncHandler) Handler_ShowAutoSyncConfig(c fiber.Ctx) error {
	configs, err := a.List()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": 1,
			"msg":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    configs,
	})
}

func (a *AutoSyncHandler) Handler_UpdateAutoSyncConfig(c fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var config autosync.SaurfangAutoSync
	if err := c.Bind().Body(&config); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": 1,
			"msg":    err.Error(),
		})
	}
	if err := a.Update(uint(id), &config); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": 1,
			"msg":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
	})
}

func (a *AutoSyncHandler) Handler_DeleteAutoSyncConfig(c fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := a.Delete(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": 1,
			"msg":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
	})
}
func (a *AutoSyncHandler) Handler_AutoSync(c fiber.Ctx) error {
	var target = struct {
		Target string `json:"target"`
	}{}
	if err := c.Bind().Body(&target); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": 1,
			"msg":    err.Error(),
		})
	}
	targetLabel := strings.Split(target.Target, "-")
	if len(targetLabel) < 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": 1,
			"msg":    target.Target + " is invalid",
		})
		if targetLabel[0] == "阿里云" {
			if err := a.AutoSyncAliYunEcs(targetLabel[1]); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"status": 1,
					"msg":    err.Error(),
				})
			}
		} else {
			if err := a.AutoSyncHuaweiECS(targetLabel[1]); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"status": 1,
					"msg":    err.Error(),
				})
			}
		}
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
	})
}
func (a *AutoSyncHandler) Handler_AutoSyncConfigSelect(c fiber.Ctx) error {
	configs, err := a.List()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": 1,
			"msg":    err.Error(),
		})
	}
	var op amis.AmisOptionsString
	var ops []amis.AmisOptionsString
	for _, sn := range *configs {
		op.Label = sn.Label
		op.SelectMode = "tree"
		op.Value = sn.Cloud + "-" + sn.Label
		ops = append(ops, op)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data": fiber.Map{
			"options": ops,
		},
	})
}
func (a *AutoSyncHandler) AutoSyncAliYunEcs(target string) error {
	var autoConfig autosync.SaurfangAutoSync
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
	var localServers []gamehost.SaurfangHosts
	if err := config.DB.Raw("select * from saurfang_hosts").Scan(&localServers).Error; err != nil {
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
					//fmt.Println("[!=]", ip(server.PublicIpAddress.IpAddress), ip(server.VpcAttributes.PrivateIpAddress.IPAddress))
					sql := fmt.Sprintf("UPDATE saurfang_hosts SET hostname = '%s',public_ip = '%s',private_ip = '%s',cpu='%s',memory='%s',os_name='%s' WHERE instance_id = '%s';",
						server.InstanceName, pkg.SingleIP(server.PublicIpAddress.IpAddress), pkg.SingleIP(server.VpcAttributes.PrivateIpAddress.IPAddress), strconv.Itoa(server.Cpu),
						strconv.Itoa(server.Memory), server.OSName, server.InstanceId)
					sqls = append(sqls, sql)
				}
				// 开始映射到本地结构体，插入或修改 update
			} else {
				sql := fmt.Sprintf("INSERT INTO saurfang_hosts (created_at, instance_id, hostname, public_ip, private_ip, cpu, memory, os_name,instance_type,port) VALUES('%v','%s','%s','%s','%s','%s','%s','%s','%s','22');", time.Now().Format("2006-01-02T15:04:05"),

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
	var autoConfig autosync.SaurfangAutoSync
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
	var localServers []gamehost.SaurfangHosts
	if err := config.DB.Raw("select * from saurfang_hosts").Scan(&localServers).Error; err != nil {
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
	var private_ip, public_ip string
	for _, cloudServer := range *cloudServers {
		for _, server := range cloudServer.Servers {
			for _, ips := range server.Addresses[server.MetaData.VpcID] {
				if ips.OSEXTIPStype == "fixed" {
					private_ip = ips.Addr
				} else if ips.OSEXTIPStype == "floating" {
					public_ip = ips.Addr
				}
			}
			if hashstring, ok := localServerMap[server.Id]; ok {
				fmt.Printf("private: ", private_ip, "public: ", public_ip)
				if pkg.Hash(fmt.Sprintf("%s-%s-%s-%s-%s-%s", server.Id, strconv.Itoa(server.Flavor.Vcpus), strconv.Itoa(server.Flavor.Ram), private_ip, public_ip, server.Name)) != hashstring {
					sql := fmt.Sprintf("UPDATE saurfang_hosts SET hostname = '%s',public_ip = '%s',private_ip = '%s',cpu='%s',memory='%s',os_name='%s' WHERE instance_id = '%s';",
						server.Name, public_ip, private_ip, strconv.Itoa(server.Flavor.Vcpus),
						strconv.Itoa(server.Flavor.Ram), server.MetaData.ImageName, server.Id)
					sqls = append(sqls, sql)
				}
				// 开始映射到本地结构体，插入或修改 update
			} else {
				sql := fmt.Sprintf("INSERT INTO saurfang_hosts (created_at, instance_id, hostname, public_ip, private_ip, cpu, memory, os_name,instance_type,port) VALUES('%v','%s','%s','%s','%s','%s','%s','%s','%s','22');", time.Now().Format("2006-01-02T15:04:05"),
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
