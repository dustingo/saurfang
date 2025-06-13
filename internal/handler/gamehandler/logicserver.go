package gamehandler

import (
	"fmt"
	"net/http"
	"saurfang/internal/models/gameserver"
	"saurfang/internal/models/serverconfig"
	"saurfang/internal/service/gameservice"
	"saurfang/internal/tools"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type LogicServerHandler struct {
	gameservice.LogicServerService
}

func NewLogicServerHandler(svc *gameservice.LogicServerService) *LogicServerHandler {
	return &LogicServerHandler{*svc}
}

// Handler_CreateLogicServer 创建游戏逻辑服
func (l *LogicServerHandler) Handler_CreateLogicServer(c *fiber.Ctx) error {
	var server gameserver.SaurfangGames
	if err := c.BodyParser(&server); err != nil {
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

// Handler_DeleteLogicServer 删除逻辑服
func (l *LogicServerHandler) Handler_DeleteLogicServer(c *fiber.Ctx) error {
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

// Handler_DeleteHostFromLogicServer 从逻辑服中删除指定的主机
func (l *LogicServerHandler) Handler_DeleteHostFromLogicServer(c *fiber.Ctx) error {
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
		if err := tx.Exec("DELETE FROM saurfang_game_hosts WHERE game_id = ? AND host_id = ?", gameid, id).Error; err != nil {
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

// Handler_UpdateLogicServer 更新逻辑服信息
func (l *LogicServerHandler) Handler_UpdateLogicServer(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var servers gameserver.SaurfangGames
	if err := c.BodyParser(&servers); err != nil {
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

// Handler_ListLogicServer 展示逻辑服信息
func (l *LogicServerHandler) Handler_ShowLogicServer(c *fiber.Ctx) error {
	servers, err := l.List()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    servers,
	})
}

// Handler_ShowServerDetail 展示逻辑服详细信息
func (l *LogicServerHandler) Handler_ShowServerDetail(c *fiber.Ctx) error {
	var gameID string
	gameID = c.Query("gameid", "0")

	id, _ := strconv.Atoi(gameID)
	var serversDetail []gameserver.GameHostsDetail
	var serverDetail gameserver.GameHostsDetail
	if id < 1 {
		err := l.DB.Raw("SELECT sh.hostname,sh.id,sh.private_ip from `saurfang_hosts` sh join `saurfang_game_hosts` sgh on sh.id = sgh.host_id  join `saurfang_games`	 sg  on sg.id  = sgh.game_id;").Scan(&serversDetail).Error
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
		if err := l.DB.Where("id = ?", id).First(&serverDetail).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  1,
				"message": err.Error(),
			})

		}
	}
	if err := l.DB.Raw("SELECT sh.hostname,sh.id,sh.private_ip from `saurfang_hosts` sh join `saurfang_game_hosts` sgh on sh.id = sgh.host_id  join `saurfang_games`	 sg  on sg.id  = sgh.game_id WHERE sg.id  = ?;", gameID).Scan(&serversDetail).Error; err != nil {
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

// Handler_ShowGameserverByTree 选择对应逻辑服
func (l *LogicServerHandler) Handler_ShowGameserverByTree(c *fiber.Ctx) error {
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
    saurfang_channels c
JOIN 
    saurfang_games g ON c.id = g.channel_id
JOIN 
    saurfang_game_hosts gh ON g.id = gh.game_id 
JOIN 
    saurfang_hosts h ON gh.host_id = h.id;`
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

// Handler_ShowServerDetailForPicker amis picker专用接口
func (l *LogicServerHandler) Handler_ShowServerDetailForPicker(c *fiber.Ctx) error {
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
    h.hostname AS host_name,
    h.private_ip
FROM 
    saurfang_channels c
JOIN 
    saurfang_games g ON c.id = g.channel_id
JOIN 
    saurfang_game_hosts gh ON g.id = gh.game_id 
JOIN 
    saurfang_hosts h ON gh.host_id = h.id;`
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

// Handler_AddHostsToLogicServer 为逻辑服分配主机
func (l *LogicServerHandler) Handler_AddHostsToLogicServer(c *fiber.Ctx) error {
	gameID, _ := strconv.Atoi(c.Query("games_id"))
	hostIds := c.Query("host_ids")
	hostID := strings.Split(hostIds, ",")
	for _, id := range hostID {
		uid, _ := strconv.Atoi(id)
		if err := l.DB.Exec("INSERT INTO saurfang_game_hosts (game_id, host_id) VALUES (?, ?)", gameID, uint(uid)).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  1,
				"message": err.Error(),
			})
		}
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"status":  0,
		"message": "success",
	})

}
