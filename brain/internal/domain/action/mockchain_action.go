package action

import (
	"github.com/sjtutwilight/Twilight/brain/internal/domain/model"
	"github.com/sjtutwilight/Twilight/brain/internal/utils"
)

type MockChainOpAction struct{}

func (m *MockChainOpAction) Execute(evt model.Event, cfg map[string]interface{}) error {
	operation, _ := cfg["operation"].(string)
	utils.Logger.Printf("Mock chain operation executed: %s", operation)
	return nil
}
