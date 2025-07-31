package tools

import "saurfang/internal/models/serverconfig"

func AddToTree(tree map[string]*serverconfig.Node, ChannelName, GameName, HostName, Private_ip string) {
	// 1. 获取或创建大类节点
	if _, exists := tree[ChannelName]; !exists {
		tree[ChannelName] = &serverconfig.Node{Label: ChannelName, SelectMode: "tree", Children: []*serverconfig.Node{}}
	}
	ChannelNameNode := tree[ChannelName]

	// 2. 检查小类节点是否已经存在于大类节点中
	var GameNameNode *serverconfig.Node
	for _, child := range ChannelNameNode.Children {
		if child.Label == GameName {
			GameNameNode = child
			break
		}
	}
	// 如果小类节点不存在，则创建它并加入大类的 Children 列表
	if GameNameNode == nil {
		GameNameNode = &serverconfig.Node{Label: GameName, SelectMode: "tree", Children: []*serverconfig.Node{}}
		ChannelNameNode.Children = append(ChannelNameNode.Children, GameNameNode)
	}

	// 3. 检查游戏节点是否已经存在于小类节点中
	var gameNode *serverconfig.Node
	for _, child := range GameNameNode.Children {
		if child.Label == HostName {
			gameNode = child
			break
		}
	}
	// 如果游戏节点不存在，则创建它并加入小类的 Children 列表
	if gameNode == nil {
		gameNode = &serverconfig.Node{Label: HostName, SelectMode: "tree", Value: Private_ip} // 设置 Value 字段为 Private_ip 地址
		GameNameNode.Children = append(GameNameNode.Children, gameNode)
	}
}

func GenerateSeverList(list map[string]*serverconfig.ServerListNode, ChannelName, GameName, serverID string) {

	// 1. 获取或创建大类节点
	if _, exists := list[ChannelName]; !exists {
		list[ChannelName] = &serverconfig.ServerListNode{Label: ChannelName, SelectMode: "list", Children: []*serverconfig.Children{}}
	}
	ChannelNameNode := list[ChannelName]

	// 2. 检查游戏名称是否已经存在于大类节点的 Children 中
	var gameExists bool
	for _, child := range ChannelNameNode.Children {
		if child.Label == GameName {
			gameExists = true
			break
		}
	}
	// 如果游戏名称不存在，则创建它并加入大类的 Children 列表
	if !gameExists {
		gameNode := &serverconfig.Children{
			Label: GameName,
			Value: serverID,
		}
		ChannelNameNode.Children = append(ChannelNameNode.Children, gameNode)
	}

	// 3. 检查服务器ID是否已经存在于大类节点的 Children 中
	var serverExists bool
	for _, child := range ChannelNameNode.Children {
		if child.Value == serverID {
			serverExists = true
			break
		}
	}
	// 如果服务器ID不存在，则创建它并加入大类的 Children 列表
	if !serverExists {
		serverNode := &serverconfig.Children{
			Label: GameName,
			Value: serverID,
		}
		ChannelNameNode.Children = append(ChannelNameNode.Children, serverNode)
	}

}
