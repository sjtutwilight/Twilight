package condition

import (
	"fmt"

	"github.com/sjtutwilight/Twilight/brain/internal/domain/model"
)

type CompositeCondition struct {
	conditions []Condition
	operator   string // "AND" or "OR"
}

func NewCompositeCondition(operator string, conditions ...Condition) (*CompositeCondition, error) {
	if operator != "AND" && operator != "OR" {
		return nil, fmt.Errorf("invalid operator: %s, must be AND or OR", operator)
	}
	return &CompositeCondition{
		conditions: conditions,
		operator:   operator,
	}, nil
}

func (c *CompositeCondition) Evaluate(event model.Event, cacheData map[string]interface{}) bool {
	if len(c.conditions) == 0 {
		return true
	}

	if c.operator == "AND" {
		for _, cond := range c.conditions {
			if !cond.Evaluate(event, cacheData) {
				return false
			}
		}
		return true
	}

	// OR operator
	for _, cond := range c.conditions {
		if cond.Evaluate(event, cacheData) {
			return true
		}
	}
	return false
}

// CreateCompositeCondition creates a composite condition from config
func CreateCompositeCondition(config map[string]interface{}) (Condition, error) {
	operator, ok := config["operator"].(string)
	if !ok {
		return nil, fmt.Errorf("operator not found in config")
	}

	conditionsConfig, ok := config["conditions"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("conditions not found in config")
	}

	var conditions []Condition
	for _, condConfig := range conditionsConfig {
		condMap, ok := condConfig.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid condition config format")
		}

		condType, ok := condMap["type"].(string)
		if !ok {
			return nil, fmt.Errorf("condition type not found")
		}

		cond, ok := GetCondition(condType)
		if !ok {
			return nil, fmt.Errorf("condition type %s not registered", condType)
		}
		conditions = append(conditions, cond)
	}

	return NewCompositeCondition(operator, conditions...)
}

func init() {
	RegisterCondition("composite", &CompositeCondition{})
}
