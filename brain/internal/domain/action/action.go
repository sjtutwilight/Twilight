package action

import "github.com/sjtutwilight/Twilight/brain/internal/domain/model"

type Action interface {
	Execute(evt model.Event, cfg map[string]interface{}) error
}
