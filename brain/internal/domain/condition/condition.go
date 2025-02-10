package condition

import "github.com/sjtutwilight/Twilight/brain/internal/domain/model"

type Condition interface {
	Evaluate(evt model.Event, nodeConfig map[string]interface{}) bool
}
