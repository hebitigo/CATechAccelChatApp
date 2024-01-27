package ws

import "github.com/google/uuid"

type channelId uuid.UUID

type serverId uuid.UUID

type broadcastMessage struct {
	serverId    serverId
	channelId   channelId
	jsonMessage []byte
}

type Hub struct {
	UserPresence map[serverId]map[channelId]map[*User]bool
	broadcast    chan broadcastMessage
	register     chan *User
	unregister   chan *User
}

func NewHub() *Hub {
	return &Hub{
		broadcast:    make(chan broadcastMessage),
		register:     make(chan *User),
		unregister:   make(chan *User),
		UserPresence: make(map[serverId]map[channelId]map[*User]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case user := <-h.register:
			//websocketで接続されたアクティブなユーザーをhubに格納するので、
			//hub上にSeverやChannelが登録されているかどうかは
			//DBに登録されているかどうかとはまた別の話で
			//DBにサーバーが登録されていたとしても、websocketで接続要求がきたuserの参加しているサーバー、チャンネル以外は
			//hubに登録されない
			if _, ok := h.UserPresence[serverId(user.ServerID)]; !ok {
				h.UserPresence[serverId(user.ServerID)] = make(map[channelId]map[*User]bool)
			}
			if _, ok := h.UserPresence[serverId(user.ServerID)][channelId(user.ChannelID)]; !ok {
				h.UserPresence[serverId(user.ServerID)][channelId(user.ChannelID)] = make(map[*User]bool)
			}

			h.UserPresence[serverId(user.ServerID)][channelId(user.ChannelID)][user] = true
		case user := <-h.unregister:
			if _, ok := h.UserPresence[serverId(user.ServerID)]; ok {
				if _, ok := h.UserPresence[serverId(user.ServerID)][channelId(user.ChannelID)]; ok {
					if _, ok := h.UserPresence[serverId(user.ServerID)][channelId(user.ChannelID)][user]; ok {
						delete(h.UserPresence[serverId(user.ServerID)][channelId(user.ChannelID)], user)
						close(user.send)
						if len(h.UserPresence[serverId(user.ServerID)][channelId(user.ChannelID)]) == 0 {
							delete(h.UserPresence[serverId(user.ServerID)], channelId(user.ChannelID))
						}
						if len(h.UserPresence[serverId(user.ServerID)]) == 0 {
							delete(h.UserPresence, serverId(user.ServerID))
						}
					}
				}
			}
		case broadcastInfo := <-h.broadcast:
			if _, ok := h.UserPresence[broadcastInfo.serverId]; ok {
				if _, ok := h.UserPresence[broadcastInfo.serverId][broadcastInfo.channelId]; ok {
					for user := range h.UserPresence[broadcastInfo.serverId][broadcastInfo.channelId] {
						select {
						case user.send <- broadcastInfo:
						//user.sendが閉じてる場合のブロッキングを防ぐためにdefaultを設定
						default:
							close(user.send)
							delete(h.UserPresence[broadcastInfo.serverId][broadcastInfo.channelId], user)
							if len(h.UserPresence[broadcastInfo.serverId][broadcastInfo.channelId]) == 0 {
								delete(h.UserPresence[broadcastInfo.serverId], broadcastInfo.channelId)
							}
						}
					}
				}
			}

		}
	}
}
