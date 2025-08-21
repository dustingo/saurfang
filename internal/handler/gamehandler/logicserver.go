package gamehandler

import (
	"fmt"
	"saurfang/internal/models/amis"
	"saurfang/internal/models/gameserver"
	"saurfang/internal/models/serverconfig"
	"saurfang/internal/repository/base"
	"saurfang/internal/tools"
	"saurfang/internal/tools/pkg"
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
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "request error", err.Error(), fiber.Map{})
	}
	if err := l.Create(&server); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to create logic server", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", nil)
}

// Handler_DeleteLogicServer 删除逻辑服 “/delete"
func (l *LogicServerHandler) Handler_DeleteLogicServer(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Query("games_id"))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "request error", err.Error(), fiber.Map{})
	}
	if err := l.Delete(uint(id)); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to delete logic server", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", nil)
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
			return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to delete host from logic server", err.Error(), fiber.Map{})
		}
	}
	if err := tx.Commit().Error; err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to delete host from logic server", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", nil)
}

// Handler_UpdateLogicServer 更新逻辑服信息 "/update/:id"
func (l *LogicServerHandler) Handler_UpdateLogicServer(c fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var servers gameserver.Games
	if err := c.Bind().Body(&servers); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "request error", err.Error(), fiber.Map{})
	}
	servers.ID = uint(id)
	if err := l.Update(servers.ID, &servers); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to update logic server", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", nil)
}

// Handler_ShowLogicServer 展示逻辑服信息 "/list" 支持条件搜索
func (l *LogicServerHandler) Handler_ShowLogicServer(c fiber.Ctx) error {
	// 获取分页参数
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.Query("perPage", "10"))
	if err != nil {
		pageSize = 10
	}

	// 获取搜索条件
	channelId, err := strconv.Atoi(c.Query("channelId", "0"))
	if err != nil {
		channelId = 1
	}
	serverId := c.Query("server_id")
	serverName := c.Query("server_name")
	status := c.Query("status")

	// 构建基础查询
	query := l.DB.Model(&gameserver.Games{}).Preload("Channel")

	// 应用搜索条件
	if channelId > 0 {
		query = query.Where("channel_id = ?", channelId)
	}

	if serverId != "" {
		query = query.Where("server_id LIKE ?", "%"+serverId+"%")
	}

	if serverName != "" {
		query = query.Where("name LIKE ?", "%"+serverName+"%")
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to count logic servers", err.Error(), fiber.Map{})
	}

	// 获取分页数据
	var data []gameserver.Games
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&data).Error; err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to list logic servers", err.Error(), fiber.Map{})
	}

	// 计算分页信息
	totalPages := (int(total) + pageSize - 1) / pageSize

	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", fiber.Map{
		"data":       data,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": totalPages,
	})
}
func (l *LogicServerHandler) Handler_ShowLogicServerTree(c fiber.Ctx) error {
	servers, err := l.List()
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to list logic server", err.Error(), fiber.Map{})
	}
	var ops []amis.AmisOptionsGeneric[string]
	var op amis.AmisOptionsGeneric[string]
	for _, sn := range servers {
		op.Label = sn.Name
		op.Value = sn.ServerID
		ops = append(ops, op)
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", fiber.Map{
		"options": ops,
	})
}

// Handler_ShowServerDetail 展示逻辑服详细信息 "/detail"
func (l *LogicServerHandler) Handler_ShowServerDetail(c fiber.Ctx) error {
	gameID := c.Query("gameid", "0")
	id, _ := strconv.Atoi(gameID)
	var serversDetail []gameserver.GameHostsDetail
	if id < 1 {
		err := l.DB.Raw("SELECT sh.hostname,sh.id,sh.private_ip from `hosts` sh join `game_hosts` sgh on sh.id = sgh.host_id  join `games`	 sg  on sg.id  = sgh.game_id;").Scan(&serversDetail).Error
		if err != nil {
			return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to show server detail", err.Error(), fiber.Map{})
		}
		return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", serversDetail)

	} else {
		if err := l.DB.Raw("SELECT sh.hostname,sh.id,sh.private_ip from `hosts` sh join `game_hosts` sgh on sh.id = sgh.host_id  join `games` sg  on sg.id  = sgh.game_id WHERE sg.id  = ?;", gameID).Scan(&serversDetail).Error; err != nil {
			return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to show server detail", err.Error(), fiber.Map{})
		}
		return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", serversDetail)
	}
}

// Handler_ShowGameserverByTree 选择对应逻辑服treeselect "/detail/select"
func (l *LogicServerHandler) Handler_ShowGameserverByTree(c fiber.Ctx) error {
	gameInfo, err := l.getGameInfo()
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to get game info", err.Error(), fiber.Map{})
	}
	nodes := make(map[string]*serverconfig.Node)
	for _, info := range gameInfo {
		tools.AddToTree(nodes, info.ChannelName, fmt.Sprintf("%s(%s)", info.GameName, info.ServerID), info.HostName, info.PrivateIP)
	}
	result := []*serverconfig.Node{}
	for _, node := range nodes {
		result = append(result, node)
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", result)
}

// Handler_ShowServerDetailForPicker amis picker专用接口 "/detail/picker"
func (l *LogicServerHandler) Handler_ShowServerDetailForPicker(c fiber.Ctx) error {
	gameInfo, err := l.getGameInfo()
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to get game info", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", gameInfo)
}

// Handler_AddHostsToLogicServer 为逻辑服分配主机 "/assignhost"
func (l *LogicServerHandler) Handler_AddHostsToLogicServer(c fiber.Ctx) error {
	gameID, _ := strconv.Atoi(c.Query("games_id"))
	hostIds := c.Query("host_ids")
	hostID := strings.Split(hostIds, ",")
	for _, id := range hostID {
		uid, _ := strconv.Atoi(id)
		if err := l.DB.Exec("INSERT INTO game_hosts (game_id, host_id) VALUES (?, ?)", gameID, uint(uid)).Error; err != nil {
			return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to add hosts to logic server", err.Error(), fiber.Map{})
		}
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", nil)

}

// Handler_TreeSelectForSyncServerConfig 发布服务器端配置文件时选择服务器 "/detail/for-config/select"
func (l *LogicServerHandler) Handler_TreeSelectForSyncServerConfig(c fiber.Ctx) error {
	gameInfo, err := l.getGameInfo()
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to get game info", err.Error(), fiber.Map{})
	}
	nodes := make(map[string]*serverconfig.Node)
	for _, info := range gameInfo {
		tools.AddToTree(nodes, info.ChannelName, fmt.Sprintf("%s(%s)", info.GameName, info.ServerID), info.HostName, fmt.Sprintf("%s-%s", info.ServerID, info.PrivateIP))
	}
	result := []*serverconfig.Node{}
	for _, node := range nodes {
		result = append(result, node)
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", result)
}

// Handler_ShowChannelServerList 执行发布任务时,选择服务器
func (l *LogicServerHandler) Handler_ShowChannelServerList(c fiber.Ctx) error {
	serverList := []struct {
		Name      string `json:"name"`
		ServerID  string `json:"server_id"`
		Channel   string `json:"channel"`
		ChannelID uint   `json:"channel_id"`
	}{}
	if err := l.DB.Debug().Table("games as g").Select("g.name,g.server_id,c.name as channel,c.id as channel_id").
		Joins("join channels c on g.channel_id = c.id").
		Scan(&serverList).Error; err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "get server list failed", err.Error(), fiber.Map{})
	}
	nodes := make(map[string]*serverconfig.ServerListNode)
	for _, info := range serverList {
		tools.GenerateSeverList(nodes, info.Channel, fmt.Sprintf("%s(%s)", info.Name, info.ServerID), info.ServerID)
	}
	result := []*serverconfig.ServerListNode{}
	for _, node := range nodes {
		result = append(result, node)
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", result)
}

func (l *LogicServerHandler) getGameInfo() ([]struct {
	ChannelName string `json:"channel_name"`
	GameName    string `json:"game_name"`
	ServerID    string `json:"server_id"`
	HostName    string `json:"host_name"`
	PrivateIP   string `json:"private_ip"`
}, error) {
	var gameInfo []struct {
		ChannelName string `json:"channel_name"`
		GameName    string `json:"game_name"`
		ServerID    string `json:"server_id"`
		HostName    string `json:"host_name"`
		PrivateIP   string `json:"private_ip"`
	}

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

	err := l.DB.Raw(query).Scan(&gameInfo).Error
	return gameInfo, err
}
