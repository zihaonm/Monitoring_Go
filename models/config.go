package models

// TelegramConfig holds Telegram bot configuration
type TelegramConfig struct {
	BotToken string `json:"bot_token" binding:"required"`
	ChatID   string `json:"chat_id" binding:"required"`
	Enabled  bool   `json:"enabled"`
}

// NotificationEvent represents different notification triggers
type NotificationEvent string

const (
	EventServiceDown NotificationEvent = "service_down"
	EventServiceUp   NotificationEvent = "service_up"
)
