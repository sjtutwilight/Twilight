package action

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sjtutwilight/Twilight/brain/internal/domain/model"
)

func TestNotificationAction_Execute(t *testing.T) {
	// 创建一个测试服务器来模拟webhook目标
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tests := []struct {
		name       string
		event      model.Event
		config     map[string]interface{}
		wantErr    bool
		errMessage string
	}{
		{
			name: "webhook notification - success",
			event: model.Event{
				ID:   "test-event",
				Type: "test",
				Data: map[string]interface{}{
					"value": 100,
				},
			},
			config: map[string]interface{}{
				"type":   "webhook",
				"url":    server.URL,
				"method": "POST",
			},
			wantErr: false,
		},
		{
			name: "webhook notification - missing url",
			event: model.Event{
				ID:   "test-event",
				Type: "test",
			},
			config: map[string]interface{}{
				"type": "webhook",
			},
			wantErr:    true,
			errMessage: "webhook url not found in config",
		},
		{
			name: "email notification",
			event: model.Event{
				ID:   "test-event",
				Type: "test",
			},
			config: map[string]interface{}{
				"type":     "email",
				"template": "test-template",
				"params": map[string]interface{}{
					"subject": "Test Subject",
				},
			},
			wantErr: false,
		},
		{
			name: "unknown notification type",
			event: model.Event{
				ID:   "test-event",
				Type: "test",
			},
			config: map[string]interface{}{
				"type": "unknown",
			},
			wantErr:    true,
			errMessage: "unknown notification type: unknown",
		},
		{
			name: "missing notification type",
			event: model.Event{
				ID:   "test-event",
				Type: "test",
			},
			config:     map[string]interface{}{},
			wantErr:    true,
			errMessage: "notification type not found in config",
		},
	}

	action := &NotificationAction{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := action.Execute(tt.event, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NotificationAction.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMessage {
				t.Errorf("NotificationAction.Execute() error message = %v, want %v", err.Error(), tt.errMessage)
			}
		})
	}
}
