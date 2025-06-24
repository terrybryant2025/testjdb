package main

// SFSObject 协议 - 消息类型枚举
const (
	Handshake                    = 0   // 握手
	Login                        = 1   // 登录
	Logout                       = 2   // 登出
	JoinRoom                     = 4   // 加入房间
	CreateRoom                   = 6   // 创建房间
	GenericMessage               = 7   // 通用消息
	ChangeRoomName               = 8   // 修改房间名称
	ChangeRoomPassword           = 9   // 修改房间密码
	SetRoomVariables             = 11  // 设置房间变量
	SetUserVariables             = 12  // 设置用户变量
	CallExtension                = 13  // 调用扩展
	LeaveRoom                    = 14  // 离开房间
	SubscribeRoomGroup           = 15  // 订阅房间组
	UnsubscribeRoomGroup         = 16  // 取消订阅房间组
	SpectatorToPlayer            = 17  // 观察者切换为玩家
	PlayerToSpectator            = 18  // 玩家切换为观察者
	ChangeRoomCapacity           = 19  // 修改房间容量
	KickUser                     = 24  // 踢出用户
	BanUser                      = 25  // 封禁用户
	FindRooms                    = 27  // 查找房间
	FindUsers                    = 28  // 查找用户
	PingPong                     = 29  // 心跳PingPong
	SetUserPosition              = 30  // 设置用户位置
	QuickJoinOrCreateRoom        = 31  // 快速加入或创建房间
	InitBuddyList                = 200 // 初始化好友列表
	AddBuddy                     = 201 // 添加好友
	BlockBuddy                   = 202 // 屏蔽好友
	RemoveBuddy                  = 203 // 移除好友
	SetBuddyVariables            = 204 // 设置好友变量
	GoOnline                     = 205 // 上线
	InviteUsers                  = 300 // 邀请用户
	InvitationReply              = 301 // 邀请回复
	CreateSFSGame                = 302 // 创建SFS游戏
	QuickJoinGame                = 303 // 快速加入游戏
	JoinRoomInvite               = 304 // 加入房间邀请
	ClusterJoinOrCreateRequest   = 500 // 集群加入或创建请求
	ClusterInviteUsers           = 502 // 集群邀请用户
	GameServerConnectionRequired = 600 // 需要游戏服务器连接
)

type TileType int

const (
	TileUnknown     TileType = iota
	TileCake                 // 1
	TilePudding              // 2
	TileOrangeCandy          // 3
	TileBlueCandy            // 4
	TileGreenCandy           // 5
	TileRedStriped           // 6
	TilePurpleCandy          // 7
	TileDrink                // 8
	TileLollipop             // 9
	TileCupcake              // 10
)
