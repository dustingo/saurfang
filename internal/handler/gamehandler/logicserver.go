package gamehandler

import (
	"saurfang/internal/models/gameserver"
	"saurfang/internal/service/gameservice"
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

// Handler_ListLogicServer
func (l *LogicServerHandler) Handler_ListLogicServer(c *fiber.Ctx) error {
	servers, err := l.List()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}

}
