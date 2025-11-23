package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"monitoring/models"
	"net/http"
	"sync"
	"time"
)

// TelegramService handles sending notifications via Telegram
type TelegramService struct {
	config *models.TelegramConfig
	mu     sync.RWMutex
	onSave func() // callback when config changes
}

// NewTelegramService creates a new Telegram notification service
func NewTelegramService() *TelegramService {
	return &TelegramService{
		config: &models.TelegramConfig{
			Enabled: false,
		},
	}
}

// SetOnSave sets the callback for when config changes
func (t *TelegramService) SetOnSave(onSave func()) {
	t.onSave = onSave
}

// LoadConfig loads configuration (used during startup)
func (t *TelegramService) LoadConfig(config *models.TelegramConfig) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if config != nil {
		t.config = config
	}
}

// GetRawConfig returns the actual config (for persistence)
func (t *TelegramService) GetRawConfig() *models.TelegramConfig {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.config
}

// SetConfig updates the Telegram configuration
func (t *TelegramService) SetConfig(config *models.TelegramConfig) {
	t.mu.Lock()
	t.config = config
	t.mu.Unlock()

	if t.onSave != nil {
		go t.onSave()
	}
}

// GetConfig returns the current Telegram configuration
func (t *TelegramService) GetConfig() *models.TelegramConfig {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Return a copy without exposing the token fully
	return &models.TelegramConfig{
		BotToken: maskToken(t.config.BotToken),
		ChatID:   t.config.ChatID,
		Enabled:  t.config.Enabled,
	}
}

// IsEnabled checks if Telegram notifications are enabled
func (t *TelegramService) IsEnabled() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.config.Enabled && t.config.BotToken != "" && t.config.ChatID != ""
}

// getEffectiveConfig returns the effective configuration for a service
// Uses service-specific overrides if provided, otherwise falls back to default
func (t *TelegramService) getEffectiveConfig(service *models.MonitoredService) (botToken, chatID string, enabled bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Start with defaults
	botToken = t.config.BotToken
	chatID = t.config.ChatID
	enabled = t.config.Enabled

	// Override with service-specific settings if provided
	if service.TelegramBotToken != "" {
		botToken = service.TelegramBotToken
	}
	if service.TelegramChatID != "" {
		chatID = service.TelegramChatID
	}
	if service.TelegramEnabled != nil {
		enabled = *service.TelegramEnabled
	}

	return botToken, chatID, enabled
}

// isEnabledForService checks if Telegram notifications are enabled for a specific service
func (t *TelegramService) isEnabledForService(service *models.MonitoredService) bool {
	botToken, chatID, enabled := t.getEffectiveConfig(service)
	return enabled && botToken != "" && chatID != ""
}

// SendServiceDownAlert sends an alert when a service goes down
func (t *TelegramService) SendServiceDownAlert(service *models.MonitoredService) error {
	if !t.isEnabledForService(service) {
		return nil
	}

	botToken, chatID, _ := t.getEffectiveConfig(service)

	message := fmt.Sprintf(
		"🔴 *Service Down Alert*\n\n"+
			"*Service:* %s\n"+
			"*URL:* %s\n"+
			"*Status:* DOWN\n"+
			"*Error:* %s\n"+
			"*Time:* %s",
		escapeMarkdown(service.Name),
		escapeMarkdown(service.URL),
		escapeMarkdown(service.ErrorMessage),
		service.LastCheck.Format("2006-01-02 15:04:05"),
	)

	return t.sendMessageWithConfig(message, botToken, chatID)
}

// SendServiceUpAlert sends an alert when a service comes back up
func (t *TelegramService) SendServiceUpAlert(service *models.MonitoredService) error {
	if !t.isEnabledForService(service) {
		return nil
	}

	botToken, chatID, _ := t.getEffectiveConfig(service)

	message := fmt.Sprintf(
		"🟢 *Service Recovered*\n\n"+
			"*Service:* %s\n"+
			"*URL:* %s\n"+
			"*Status:* UP\n"+
			"*Response Time:* %dms\n"+
			"*Time:* %s",
		escapeMarkdown(service.Name),
		escapeMarkdown(service.URL),
		service.ResponseTime,
		service.LastCheck.Format("2006-01-02 15:04:05"),
	)

	return t.sendMessageWithConfig(message, botToken, chatID)
}

// SendSSLExpiryAlert sends an alert when SSL certificate is expiring soon
func (t *TelegramService) SendSSLExpiryAlert(service *models.MonitoredService) error {
	if !t.isEnabledForService(service) {
		return nil
	}

	botToken, chatID, _ := t.getEffectiveConfig(service)

	message := fmt.Sprintf(
		"🔒 *SSL Certificate Expiry Alert*\n\n"+
			"*Service:* %s\n"+
			"*URL:* %s\n"+
			"*Days Until Expiry:* %d\n"+
			"*Expiry Date:* %s\n"+
			"*Issuer:* %s\n"+
			"*Time:* %s\n\n"+
			"_Please renew your SSL certificate\\\\._",
		escapeMarkdown(service.Name),
		escapeMarkdown(service.URL),
		service.SSLDaysLeft,
		service.SSLCertExpiry.Format("2006-01-02 15:04:05"),
		escapeMarkdown(service.SSLCertIssuer),
		time.Now().Format("2006-01-02 15:04:05"),
	)

	return t.sendMessageWithConfig(message, botToken, chatID)
}

// SendSystemAlert sends a system resource alert
func (t *TelegramService) SendSystemAlert(resourceType string, device string, currentValue float64, threshold float64) error {
	if !t.IsEnabled() {
		return nil
	}

	var emoji, resourceName, deviceInfo string

	switch resourceType {
	case "disk":
		emoji = "💾"
		resourceName = "Disk Space"
		deviceInfo = fmt.Sprintf("\n*Device:* %s", escapeMarkdown(device))
	case "cpu":
		emoji = "🔥"
		resourceName = "CPU Usage"
	case "memory":
		emoji = "⚠️"
		resourceName = "Memory Usage"
	default:
		emoji = "⚠️"
		resourceName = "System Resource"
	}

	message := fmt.Sprintf(
		"%s *%s Alert*\n\n"+
			"*Resource:* %s%s\n"+
			"*Current Usage:* %.1f%%\n"+
			"*Threshold:* %.1f%%\n"+
			"*Time:* %s\n\n"+
			"_Please check your system resources\\._",
		emoji,
		resourceName,
		resourceName,
		deviceInfo,
		currentValue,
		threshold,
		escapeMarkdown(time.Now().Format("2006-01-02 15:04:05")),
	)

	return t.sendMessage(message)
}

// SendTestMessage sends a test notification
func (t *TelegramService) SendTestMessage() error {
	t.mu.RLock()
	botToken := t.config.BotToken
	chatID := t.config.ChatID
	t.mu.RUnlock()

	if botToken == "" || chatID == "" {
		return fmt.Errorf("bot token and chat ID are required")
	}

	message := "✅ *Test Notification*\n\nYour monitoring service is successfully connected to Telegram!"

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       message,
		"parse_mode": "Markdown",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status code: %d", resp.StatusCode)
	}

	return nil
}

// sendMessage sends a message to Telegram using the default config
func (t *TelegramService) sendMessage(message string) error {
	t.mu.RLock()
	botToken := t.config.BotToken
	chatID := t.config.ChatID
	t.mu.RUnlock()

	return t.sendMessageWithConfig(message, botToken, chatID)
}

// sendMessageWithConfig sends a message to Telegram with specific bot token and chat ID
func (t *TelegramService) sendMessageWithConfig(message, botToken, chatID string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       message,
		"parse_mode": "Markdown",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status code: %d", resp.StatusCode)
	}

	return nil
}

// escapeMarkdown escapes special characters for Telegram markdown
func escapeMarkdown(text string) string {
	replacer := []string{
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	}

	result := text
	for i := 0; i < len(replacer); i += 2 {
		result = replaceAll(result, replacer[i], replacer[i+1])
	}
	return result
}

func replaceAll(s, old, new string) string {
	result := ""
	for i := 0; i < len(s); i++ {
		if i <= len(s)-len(old) && s[i:i+len(old)] == old {
			result += new
			i += len(old) - 1
		} else {
			result += string(s[i])
		}
	}
	return result
}

// maskToken masks the bot token for security
func maskToken(token string) string {
	if token == "" {
		return ""
	}
	if len(token) < 10 {
		return "***"
	}
	return token[:5] + "..." + token[len(token)-5:]
}
