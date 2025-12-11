package notificationbus

import (
	"context"
	"time"
)

// ScheduleConfig holds scheduling configuration for notifications.
type ScheduleConfig struct {
	MorningTime string         // HH:MM format
	EveningTime string         // HH:MM format
	Location    *time.Location // Timezone
}

// TelegramSender interface for dependency injection.
type TelegramSender interface {
	SendMessage(ctx context.Context, chatID int64, content string) (TelegramSendResponse, error)
}

// TelegramSendResponse represents the response from sending a Telegram message.
type TelegramSendResponse struct {
	MessageID int64
}
