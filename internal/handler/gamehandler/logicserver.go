package gamehandler

import (
	"fmt"
	nomadapi "github.com/hashicorp/nomad/api"
	"log"
	"log/slog"
	"net/http"
	"os"
	"saurfang/internal/models/amis"
	"saurfang/internal/models/gameserver"
	"saurfang/internal/models/serverconfig"
	"saurfang/internal/repository/base"
	"saurfang/internal/tools"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type LogicServerHandler struct {
	base.BaseGormRepository[gameserver.Games]
	base.NomadJobRepository
}

// Handler_CreateLogicServer 创建游戏逻辑服
func (l *LogicServerHandler) Handler_CreateLogicServer(c fiber.Ctx) error {
	var server gameserver.Games
	if err := c.Bind().Body(&server); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	if err := l.Create(&server); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
	})
}

// Handler_DeleteLogicServer 删除逻辑服 “/delete"
func (l *LogicServerHandler) Handler_DeleteLogicServer(c fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Query("games_id"))
	if err := l.Delete(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
	})
}

// Handler_DeleteHostFromLogicServer 从逻辑服中删除指定的主机 "/deletehosts"
func (l *LogicServerHandler) Handler_DeleteHostFromLogicServer(c fiber.Ctx) error {
	gameid, _ := strconv.Atoi(c.Query("games_id"))
	hostid := strings.Split(c.Query("host_ids"), ",")
	ids := make([]uint, 0, len(hostid))
	for _, i := range hostid {
		id, _ := strconv.Atoi(i)
		ids = append(ids, uint(id))
	}
	tx := l.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	for _, id := range ids {
		if err := tx.Exec("DELETE FROM game_hosts WHERE game_id = ? AND host_id = ?", gameid, id).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
	})
}

// Handler_UpdateLogicServer 更新逻辑服信息 "/update/:id"
func (l *LogicServerHandler) Handler_UpdateLogicServer(c fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var servers gameserver.Games
	if err := c.Bind().Body(&servers); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	servers.ID = uint(id)
	if err := l.Update(servers.ID, &servers); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
	})
}

// Handler_ListLogicServer 展示逻辑服信息 "/list"
func (l *LogicServerHandler) Handler_ShowLogicServer(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("perPage", "10"))
	channelId, _ := strconv.Atoi(c.Query("channelId", "1"))
	var data []gameserver.Games
	var total int64
	if err := l.DB.Model(&data).Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	if err := l.DB.Offset((page-1)*pageSize).Limit(pageSize).Find(&data, "channel_id = ?", channelId).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    data,
	})
}
func (l *LogicServerHandler) Handler_ShowLogicServerTree(c fiber.Ctx) error {
	servers, err := l.List()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	var ops []amis.AmisOptionsGeneric[string]
	var op amis.AmisOptionsGeneric[string]
	for _, sn := range *servers {
		op.Label = sn.Name
		op.Value = sn.ServerID
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

// Handler_ShowServerDetail 展示逻辑服详细信息 "/detail"
func (l *LogicServerHandler) Handler_ShowServerDetail(c fiber.Ctx) error {
	var gameID string
	gameID = c.Query("gameid", "0")
	id, _ := strconv.Atoi(gameID)
	var serversDetail []gameserver.GameHostsDetail
	if id < 1 {
		err := l.DB.Raw("SELECT sh.hostname,sh.id,sh.private_ip from `hosts` sh join `game_hosts` sgh on sh.id = sgh.host_id  join `games`	 sg  on sg.id  = sgh.game_id;").Scan(&serversDetail).Error
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  1,
				"message": err.Error(),
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  0,
			"message": "success",
			"data":    serversDetail,
		})

	} else {
		if err := l.DB.Raw("SELECT sh.hostname,sh.id,sh.private_ip from `hosts` sh join `game_hosts` sgh on sh.id = sgh.host_id  join `games` sg  on sg.id  = sgh.game_id WHERE sg.id  = ?;", gameID).Scan(&serversDetail).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  1,
				"message": err.Error(),
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  0,
			"message": "success",
			"data":    serversDetail,
		})
	}
}

// Handler_ShowGameserverByTree 选择对应逻辑服treeselect "/detail/select"
func (l *LogicServerHandler) Handler_ShowGameserverByTree(c fiber.Ctx) error {
	gameInfo := []struct {
		ChannelName string `json:"channel_name"`
		GameName    string `json:"game_name"`
		ServerID    string `json:"server_id"`
		HostName    string `json:"host_name"`
		PrivateIP   string `json:"private_ip"`
	}{}
	query := `SELECT 
    c.name AS channel_name,
    g.name AS game_name,
    g.server_id AS server_id,
    h.hostname AS host_name,
    h.private_ip
FROM 
    channels c
JOIN 
    games g ON c.id = g.channel_id
JOIN 
    game_hosts gh ON g.id = gh.game_id 
JOIN 
    hosts h ON gh.host_id = h.id;`
	if err := l.DB.Raw(query).Scan(&gameInfo).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	nodes := make(map[string]*serverconfig.Node)
	for _, info := range gameInfo {
		tools.AddToTree(nodes, info.ChannelName, fmt.Sprintf("%s(%s)", info.GameName, info.ServerID), info.HostName, info.PrivateIP)
	}
	result := []*serverconfig.Node{}
	for _, node := range nodes {
		result = append(result, node)
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    result,
	})
}

// Handler_ShowServerDetailForPicker amis picker专用接口 "/detail/picker"
func (l *LogicServerHandler) Handler_ShowServerDetailForPicker(c fiber.Ctx) error {
	gameInfo := []struct {
		ChannelName string `json:"channel_name"`
		GameName    string `json:"game_name"`
		ServerID    string `json:"server_id"`
		HostName    string `json:"host_name"`
		PrivateIP   string `json:"private_ip"`
	}{}
	query := `SELECT 
    c.name AS channel_name,
    g.name AS game_name,
    g.server_id AS server_id,
    h.hostname AS host_name,
    h.private_ip
FROM 
    channels c
JOIN 
    games g ON c.id = g.channel_id
JOIN 
    game_hosts gh ON g.id = gh.game_id 
JOIN 
    hosts h ON gh.host_id = h.id;`
	if err := l.DB.Raw(query).Scan(&gameInfo).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    gameInfo,
	})
}

// Handler_AddHostsToLogicServer 为逻辑服分配主机 "/assignhost"
func (l *LogicServerHandler) Handler_AddHostsToLogicServer(c fiber.Ctx) error {
	gameID, _ := strconv.Atoi(c.Query("games_id"))
	hostIds := c.Query("host_ids")
	hostID := strings.Split(hostIds, ",")
	for _, id := range hostID {
		uid, _ := strconv.Atoi(id)
		if err := l.DB.Exec("INSERT INTO game_hosts (game_id, host_id) VALUES (?, ?)", gameID, uint(uid)).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  1,
				"message": err.Error(),
			})
		}
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
	})

}

// Handler_TreeSelectForSyncServerConfig 发布服务器端配置文件时选择服务器 "/detail/for-config/select"
func (l *LogicServerHandler) Handler_TreeSelectForSyncServerConfig(c fiber.Ctx) error {
	gameInfo := []struct {
		ChannelName string `json:"channel_name"`
		GameName    string `json:"game_name"`
		ServerID    string `json:"server_id"`
		HostName    string `json:"host_name"`
		PrivateIP   string `json:"private_ip"`
	}{}
	query := `SELECT 
    c.name AS channel_name,
    g.name AS game_name,
    g.server_id AS server_id,
    h.hostname AS host_name,
    h.private_ip
FROM 
    channels c
JOIN 
    games g ON c.id = g.channel_id
JOIN 
    game_hosts gh ON g.id = gh.game_id 
JOIN 
    hosts h ON gh.host_id = h.id;`
	if err := l.DB.Raw(query).Scan(&gameInfo).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	nodes := make(map[string]*serverconfig.Node)
	for _, info := range gameInfo {
		tools.AddToTree(nodes, info.ChannelName, fmt.Sprintf("%s(%s)", info.GameName, info.ServerID), info.HostName, fmt.Sprintf("%s-%s", info.ServerID, info.PrivateIP))
	}
	result := []*serverconfig.Node{}
	for _, node := range nodes {
		result = append(result, node)
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    result,
	})
}

// Handler_ShowChannelServerList 执行发布任务时,选择服务器
func (l *LogicServerHandler) Handler_ShowChannelServerList(c fiber.Ctx) error {
	serverList := []struct {
		Name      string `json:"name"`
		ServerID  string `json:"server_id"`
		Channel   string `json:"channel"`
		ChannelID uint   `json:"channel_id"`
	}{}
	if err := l.DB.Table("games as g").Select("g.name,g.server_id,c.name as channel,c.id as channel_id").
		Joins("join channels c on g.id = c.id").
		Scan(&serverList).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"error":   err.Error(),
			"message": "get server list failed",
		})
	}
	fmt.Println("serverList = ", serverList)
	nodes := make(map[string]*serverconfig.ServerListNode)
	for _, info := range serverList {
		tools.GenerateSeverList(nodes, info.Channel, fmt.Sprintf("%s(%s)", info.Name, info.ServerID), info.ServerID)
	}
	for k, v := range nodes {
		fmt.Println(k, v)
	}
	result := []*serverconfig.ServerListNode{}
	for _, node := range nodes {
		result = append(result, node)
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    result,
	})
}

// Handler_ExecGameops 执行有媳妇操作 start|stop
//func (l *LogicServerHandler) Handler_ExecGameops(c fiber.Ctx) error {
//	serverID := c.Query("serverid")
//	ops := c.Query("ops")
//	svc := c.Query("svc")
//	config, err := l.Etcd.Get(context.Background(), tools.AddNamespace(serverID, os.Getenv("GAME_CONFIG_NAMESPACE")))
//	if err != nil {
//		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
//			"status":  1,
//			"message": err.Error(),
//		})
//	}
//	if len(config.Kvs) == 0 {
//		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
//			"status":  1,
//			"message": "key not found",
//		})
//	}
//	var configs map[string]serverconfig.Configs = make(map[string]serverconfig.Configs)
//	if svc == "" {
//		for _, kv := range config.Kvs {
//			var gameconfigs serverconfig.GameConfigs
//			if err := json.Unmarshal(kv.Value, &gameconfigs); err != nil {
//				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
//					"status":  1,
//					"message": err.Error(),
//				})
//			}
//			for _, cnf := range gameconfigs.Configs {
//				configs[fmt.Sprintf("%s-%s", serverID, cnf.SvcName)] = cnf
//			}
//		}
//	} else {
//		for _, kv := range config.Kvs {
//			var gameconfigs serverconfig.GameConfigs
//			if err := json.Unmarshal(kv.Value, &gameconfigs); err != nil {
//				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
//					"status":  1,
//					"message": err.Error(),
//				})
//			}
//			for _, cnf := range gameconfigs.Configs {
//				if cnf.SvcName == svc {
//					configs[fmt.Sprintf("%s-%s", serverID, cnf.SvcName)] = cnf
//				}
//			}
//		}
//	}
//	tools.ConcurrentExec(configs, c, ops, serverID)
//	return c.SendStatus(http.StatusOK)
//}

func (l *LogicServerHandler) Handler_Execgameops(c fiber.Ctx) error {
	serverIDs := c.Query("server_ids")
	ops := c.Query("ops")
	log.Println("BatchExecgameops", "serverids", serverIDs, "ops", ops)
	switch ops {
	case "start":
		for _, serverID := range strings.Split(serverIDs, ",") {
			gameConfig, err := l.ShowNomadJobByKey(tools.AddNamespace(serverID, os.Getenv("GAME_CONFIG_NAMESPACE")))
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"status":  1,
					"message": fmt.Sprintln("search nomad job failed,", err.Error()),
				})
			}
			job, err := l.Nomad.Jobs().ParseHCL(gameConfig.Setting, true)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"status":  1,
					"message": fmt.Sprintln("parse nomad job failed,", err.Error()),
				})
			}
			go func(id string) {
				res, _, err := l.Nomad.Jobs().Register(job, &nomadapi.WriteOptions{})
				if err != nil {
					// 操作失败,不修改服务器状态
					slog.Error("register nomad job failed", "message", err.Error(), "serverID", id, "evalID", res.EvalID, "LastContact", res.LastContact)
					return
				}
				// status == 1 online, status ==0 offline
				l.DB.Exec("UPDATE games set status = 1 where server_id = ?;", id)
			}(serverID)
		}
	case "stop":
		for _, serverID := range strings.Split(serverIDs, ",") {
			log.Println("got serverid: ", serverID)
			gameConfig, err := l.ShowNomadJobByKey(tools.AddNamespace(serverID, os.Getenv("GAME_CONFIG_NAMESPACE")))
			if err != nil {
				log.Println("errr", err.Error())
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"status":  1,
					"message": fmt.Sprintln("search nomad job failed,", err.Error()),
				})
			}
			job, err := l.Nomad.Jobs().ParseHCL(gameConfig.Setting, true)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"status":  1,
					"message": fmt.Sprintln("parse nomad job failed,", err.Error()),
				})
			}
			go func(jobID, serverID string) {
				res, _, err := l.Nomad.Jobs().Deregister(jobID, false, &nomadapi.WriteOptions{})
				if err != nil {
					// 操作失败,不修改服务器状态
					slog.Error("register nomad job failed", "message", err.Error(), "serverID", serverID, "result", res)
					return
				}
				// status == 1 online, status ==0 offline
				l.DB.Exec("UPDATE games set status = 0 where server_id = ?;", serverID)
			}(*job.ID, serverID)
		}
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "applied successfully",
	})
}
