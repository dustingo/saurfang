package boardhandler

import (
	"github.com/gofiber/fiber/v3"
	"saurfang/internal/config"
	"saurfang/internal/models/dashboard"
	"saurfang/internal/models/gamechannel"
	"saurfang/internal/models/gamegroup"
	"saurfang/internal/models/gamehost"
	"saurfang/internal/models/gameserver"
	"saurfang/internal/models/user"
	"sync"
)

func Handler_TotalResource(c fiber.Ctx) error {
	var channelCount, hostsCount, groupsCount, gamesCount, usersCount int64
	var wg sync.WaitGroup
	wg.Add(5)
	go func() {
		defer wg.Done()
		config.DB.Model(&gamechannel.Channels{}).Count(&channelCount)
	}()
	go func() {
		defer wg.Done()
		config.DB.Model(&gamehost.Hosts{}).Count(&hostsCount)
	}()
	go func() {
		defer wg.Done()
		config.DB.Model(&gamegroup.Groups{}).Count(&groupsCount)
	}()
	go func() {
		defer wg.Done()
		config.DB.Model(&gameserver.Games{}).Count(&gamesCount)
	}()
	go func() {
		defer wg.Done()
		config.DB.Model(&user.User{}).Count(&usersCount)
	}()
	wg.Wait()
	data := dashboard.ResourceStatistics{
		Channels: channelCount,
		Hosts:    hostsCount,
		Groups:   groupsCount,
		Games:    gamesCount,
		Users:    usersCount,
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    data,
	})
}
