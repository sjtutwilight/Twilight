package condition

import (
	"testing"

	"github.com/sjtutwilight/Twilight/brain/internal/domain/model"
)

func TestThresholdCondition_Evaluate(t *testing.T) {
	tests := []struct {
		name      string
		condition ThresholdCondition
		event     model.Event
		context   map[string]interface{}
		want      bool
	}{
		{
			name: "simple greater than - true",
			condition: ThresholdCondition{
				Field:     "data.value",
				Operator:  "gt",
				Threshold: 100.0,
			},
			event: model.Event{
				Data: map[string]interface{}{
					"value": 150.0,
				},
			},
			want: true,
		},
		{
			name: "simple greater than - false",
			condition: ThresholdCondition{
				Field:     "data.value",
				Operator:  "gt",
				Threshold: 100.0,
			},
			event: model.Event{
				Data: map[string]interface{}{
					"value": 50.0,
				},
			},
			want: false,
		},
		{
			name: "nested field",
			condition: ThresholdCondition{
				Field:     "data.metrics.value",
				Operator:  "gt",
				Threshold: 100.0,
			},
			event: model.Event{
				Data: map[string]interface{}{
					"metrics": map[string]interface{}{
						"value": 150.0,
					},
				},
			},
			want: true,
		},
		{
			name: "field not found",
			condition: ThresholdCondition{
				Field:     "data.nonexistent",
				Operator:  "gt",
				Threshold: 100.0,
			},
			event: model.Event{
				Data: map[string]interface{}{
					"value": 150.0,
				},
			},
			want: false,
		},
		{
			name: "invalid operator",
			condition: ThresholdCondition{
				Field:     "data.value",
				Operator:  "invalid",
				Threshold: 100.0,
			},
			event: model.Event{
				Data: map[string]interface{}{
					"value": 150.0,
				},
			},
			want: false,
		},
		{
			name: "type conversion - int to float",
			condition: ThresholdCondition{
				Field:     "data.value",
				Operator:  "gt",
				Threshold: 100.0,
			},
			event: model.Event{
				Data: map[string]interface{}{
					"value": 150,
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.condition.Evaluate(tt.event, tt.context); got != tt.want {
				t.Errorf("ThresholdCondition.Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}
