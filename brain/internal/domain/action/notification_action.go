package action

import (
	"fmt"

	"github.com/sjtutwilight/Twilight/brain/internal/domain/model"
	"github.com/sjtutwilight/Twilight/brain/internal/utils"
)

type NotificationAction struct{}

func (a *NotificationAction) Execute(event model.Event, config map[string]interface{}) error {
	notificationType, ok := config["type"].(string)
	if !ok {
		return fmt.Errorf("notification type not found in config")
	}

	switch notificationType {
	case "webhook":
		return mockWebhook(event, config)
	case "chain":
		return mockChainExecution(event, config)
	case "user":
		return mockUserNotification(event, config)
	default:
		return fmt.Errorf("unknown notification type: %s", notificationType)
	}
}

func mockWebhook(event model.Event, config map[string]interface{}) error {
	url, ok := config["url"].(string)
	if !ok {
		return fmt.Errorf("webhook url not found in config")
	}

	utils.Logger.Printf("[MOCK] Webhook notification sent to %s with event ID: %s", url, event.ID)
	return nil
}

func mockChainExecution(event model.Event, config map[string]interface{}) error {
	contract, ok := config["contract"].(string)
	if !ok {
		return fmt.Errorf("contract address not found in config")
	}

	method, ok := config["method"].(string)
	if !ok {
		return fmt.Errorf("contract method not found in config")
	}

	utils.Logger.Printf("[MOCK] Chain execution on contract %s, method %s with event ID: %s", contract, method, event.ID)
	return nil
}

func mockUserNotification(event model.Event, config map[string]interface{}) error {
	userID, ok := config["user_id"].(string)
	if !ok {
		return fmt.Errorf("user_id not found in config")
	}

	channel, ok := config["channel"].(string)
	if !ok {
		channel = "default"
	}

	utils.Logger.Printf("[MOCK] User notification sent to user %s via %s channel with event ID: %s", userID, channel, event.ID)
	return nil
}

func init() {
	RegisterAction("notification", &NotificationAction{})
}
