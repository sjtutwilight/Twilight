package condition

import (
	"reflect"
	"strings"

	"github.com/sjtutwilight/Twilight/brain/internal/domain/model"
)

type ThresholdCondition struct {
	Field     string      `json:"field"`
	Operator  string      `json:"operator"`
	Threshold interface{} `json:"threshold"`
}

func (c *ThresholdCondition) Evaluate(event model.Event, context map[string]interface{}) bool {
	// 获取字段值
	value := getFieldValue(event, c.Field)
	if value == nil {
		return false
	}

	// 比较值
	return compare(value, c.Threshold, c.Operator)
}

// getFieldValue 从事件中获取指定字段的值，支持嵌套字段（如 "data.value"）
func getFieldValue(event model.Event, field string) interface{} {
	parts := strings.Split(field, ".")
	var current interface{} = map[string]interface{}{
		"data":    event.Data,
		"context": event.Context,
	}

	for _, part := range parts {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[part]
		} else {
			return nil
		}
	}

	return current
}

// compare 比较两个值
func compare(value, threshold interface{}, operator string) bool {
	v1 := reflect.ValueOf(value)
	v2 := reflect.ValueOf(threshold)

	// 如果类型不同，尝试转换
	if v1.Type() != v2.Type() {
		switch v2.Kind() {
		case reflect.Float64:
			v1f, ok := toFloat64(value)
			if !ok {
				return false
			}
			v2f := v2.Float()
			return compareFloat64(v1f, v2f, operator)
		case reflect.Int64:
			v1i, ok := toInt64(value)
			if !ok {
				return false
			}
			v2i := v2.Int()
			return compareInt64(v1i, v2i, operator)
		default:
			return false
		}
	}

	// 类型相同，直接比较
	switch v1.Kind() {
	case reflect.Float64:
		return compareFloat64(v1.Float(), v2.Float(), operator)
	case reflect.Int64:
		return compareInt64(v1.Int(), v2.Int(), operator)
	default:
		return false
	}
}

func compareFloat64(v1, v2 float64, operator string) bool {
	switch operator {
	case "gt":
		return v1 > v2
	case "gte":
		return v1 >= v2
	case "lt":
		return v1 < v2
	case "lte":
		return v1 <= v2
	case "eq":
		return v1 == v2
	default:
		return false
	}
}

func compareInt64(v1, v2 int64, operator string) bool {
	switch operator {
	case "gt":
		return v1 > v2
	case "gte":
		return v1 >= v2
	case "lt":
		return v1 < v2
	case "lte":
		return v1 <= v2
	case "eq":
		return v1 == v2
	default:
		return false
	}
}

func toFloat64(v interface{}) (float64, bool) {
	switch v := v.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

func toInt64(v interface{}) (int64, bool) {
	switch v := v.(type) {
	case int64:
		return v, true
	case int:
		return int64(v), true
	case float64:
		return int64(v), true
	case float32:
		return int64(v), true
	default:
		return 0, false
	}
}

func init() {
	RegisterCondition("threshold", &ThresholdCondition{})
}
