package tools

import (
	"saurfang/internal/models/serverconfig"
	"saurfang/internal/testutils"
	"testing"
)

func TestGenerateSeverList(t *testing.T) {
	mockDB := testutils.SetupMockDB(t)
	defer mockDB.Close()
	//select g.name,g.server_id,c.name as channel,c.id as channel_id from games g join channels c on g.channel_id = c.id;
	serverList := []struct {
		Name      string `json:"name"`
		ServerID  string `json:"server_id"`
		Channel   string `json:"channel"`
		ChannelID uint   `json:"channel_id"`
	}{
		{
			Name:      "无与伦比",
			ServerID:  "10001",
			Channel:   "官服",
			ChannelID: 2,
		},
		{
			Name:      " 叱诧风云",
			ServerID:  "20001",
			Channel:   "硬核",
			ChannelID: 1,
		},
		{
			Name:      " 千星之城",
			ServerID:  "10002",
			Channel:   "官服",
			ChannelID: 2,
		},
	}
	nodes := make(map[string]*serverconfig.ServerListNode)
	t.Run("GenerateSeverList", func(t *testing.T) {
		for _, v := range serverList {
			GenerateSeverList(nodes, v.Channel, v.Name, v.ServerID)
		}

		// 构造期望的结果进行断言
		expected := map[string]*serverconfig.ServerListNode{
			"官服": {
				Label:      "官服",
				SelectMode: "list",
				Children: []*serverconfig.Children{
					{Label: "无与伦比", Value: "10001"},
					{Label: " 千星之城", Value: "10002"},
				},
			},
			"硬核": {
				Label:      "硬核",
				SelectMode: "list",
				Children: []*serverconfig.Children{
					{Label: " 叱诧风云", Value: "20001"},
				},
			},
		}

		// 验证nodes的长度
		if len(nodes) != len(expected) {
			t.Errorf("Expected %d channels, got %d", len(expected), len(nodes))
		}

		// 验证每个channel的内容
		for channelName, expectedNode := range expected {
			actualNode, exists := nodes[channelName]
			if !exists {
				t.Errorf("Expected channel %s not found", channelName)
				continue
			}

			// 验证基本属性
			if actualNode.Label != expectedNode.Label {
				t.Errorf("Channel %s: expected label %s, got %s", channelName, expectedNode.Label, actualNode.Label)
			}
			if actualNode.SelectMode != expectedNode.SelectMode {
				t.Errorf("Channel %s: expected selectMode %s, got %s", channelName, expectedNode.SelectMode, actualNode.SelectMode)
			}

			// 验证children数量
			if len(actualNode.Children) != len(expectedNode.Children) {
				t.Errorf("Channel %s: expected %d children, got %d", channelName, len(expectedNode.Children), len(actualNode.Children))
				continue
			}

			// 验证每个child的内容
			for i, expectedChild := range expectedNode.Children {
				actualChild := actualNode.Children[i]
				if actualChild.Label != expectedChild.Label {
					t.Errorf("Channel %s, child %d: expected label %s, got %s", channelName, i, expectedChild.Label, actualChild.Label)
				}
				if actualChild.Value != expectedChild.Value {
					t.Errorf("Channel %s, child %d: expected value %s, got %s", channelName, i, expectedChild.Value, actualChild.Value)
				}
			}
		}
	})

}
