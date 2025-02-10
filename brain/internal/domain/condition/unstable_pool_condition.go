package condition

import "github.com/sjtutwilight/Twilight/brain/internal/domain/model"

type UnstablePoolCondition struct{}

func (u *UnstablePoolCondition) Evaluate(evt model.Event, cfg map[string]interface{}) bool {
	prevPrice, _ := evt.Data["previousPrice"].(float64)
	currPrice, _ := evt.Data["price"].(float64)
	threshold := 3.0
	if th, ok := cfg["threshold"].(float64); ok {
		threshold = th
	}
	return (prevPrice - currPrice) > threshold
}

func init() {
	RegisterCondition("checkIfUnstablePool", &UnstablePoolCondition{})
}
