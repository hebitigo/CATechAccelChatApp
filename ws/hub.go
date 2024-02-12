package ws

type userId string

type Hub struct {
	UserPresence map[userId]*User
	broadcast    chan []byte
	register     chan *User
	unregister   chan *User
}

func NewHub() *Hub {
	return &Hub{
		broadcast:    make(chan []byte),
		register:     make(chan *User),
		unregister:   make(chan *User),
		UserPresence: make(map[userId]*User),
	}
}

// userが参加しているサーバー、チャンネルはメッセージには含むがbackend側ではその情報を元に
// broadcastするサーバー、チャンネルを絞る混む事はしない
// backend側ではuserがオンラインかどうか、websocketで接続しているかどうかをHubで管理し、
// なんらかの情報がフロントエンドのwebscoketから送られてきた場合には、
// Hubに登録されている全てのuserに対してbroadcastする
func (h *Hub) Run() {
	for {
		select {
		case user := <-h.register:
			h.UserPresence[userId(user.UserID)] = user
		case user := <-h.unregister:
			if _, ok := h.UserPresence[userId(user.UserID)]; ok {
				delete(h.UserPresence, userId(user.UserID))
				close(user.send)
			}
		case broadcastInfo := <-h.broadcast:
			for _, user := range h.UserPresence {
				select {
				case user.send <- broadcastInfo:
				//user.sendが閉じてる場合のブロッキングを防ぐためにdefaultを設定
				default:
					close(user.send)
					delete(h.UserPresence, userId(user.UserID))
				}
			}

		}
	}
}
