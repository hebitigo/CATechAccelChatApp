package entity

type BotEndpoint struct {
	ID       string `json:"id" bun:"id,pk,autoincrement"`
	ENDPOINT string `json:"endpoint" bun:"endpoint"`
	NAME     string `json:"name" bun:"name"`
	ICON_URL string `json:"icon_url" bun:"icon_url"`
}
