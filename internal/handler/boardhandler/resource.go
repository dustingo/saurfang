package boardhandler

import (
	"saurfang/internal/config"
	"saurfang/internal/models/dashboard"
	"saurfang/internal/models/gamechannel"
	"saurfang/internal/models/gamegroup"
	"saurfang/internal/models/gamehost"
	"saurfang/internal/models/gameserver"
	"saurfang/internal/models/user"

	"github.com/gofiber/fiber/v3"
)

func Handler_TotalResource(c fiber.Ctx) error {
	var stats dashboard.ResourceStatistics

	// 统计渠道数量
	config.DB.Model(&gamechannel.Channels{}).Count(&stats.Channels)

	// 统计主机数量
	config.DB.Model(&gamehost.Hosts{}).Count(&stats.Hosts)

	// 统计游戏组数量
	config.DB.Model(&gamegroup.Groups{}).Count(&stats.Groups)

	// 统计游戏服务器数量
	config.DB.Model(&gameserver.Games{}).Count(&stats.Games)

	// 统计用户数量
	config.DB.Model(&user.User{}).Count(&stats.Users)

	// 返回完整的ECharts配置格式
	chartData := map[string]interface{}{
		"xAxis": map[string]interface{}{
			"type": "category",
			"data": []string{"渠道", "主机", "游戏组", "游戏服务器", "用户"},
		},
		"yAxis": map[string]interface{}{
			"type": "value",
		},
		"series": []map[string]interface{}{
			{
				"name": "数量",
				"type": "bar",
				"data": []int64{stats.Channels, stats.Hosts, stats.Groups, stats.Games, stats.Users},
				"itemStyle": map[string]interface{}{
					"normal": map[string]interface{}{
						"label": map[string]interface{}{
							"show":     true,
							"position": "top",
							"textStyle": map[string]interface{}{
								"color":    "black",
								"fontSize": 16,
							},
						},
					},
				},
			},
		},
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    chartData,
	})
}
