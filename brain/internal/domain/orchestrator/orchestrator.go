package orchestrator

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sjtutwilight/Twilight/brain/internal/domain/action"
	"github.com/sjtutwilight/Twilight/brain/internal/domain/condition"
	"github.com/sjtutwilight/Twilight/brain/internal/domain/model"
	"github.com/sjtutwilight/Twilight/brain/internal/domain/ports"
	"github.com/sjtutwilight/Twilight/brain/internal/utils"
)

type Orchestrator struct {
	Strategy    *model.Strategy
	Producer    ports.Producer
	DelayTopic  string
	ResumeTopic string
}

type ResumeMessage struct {
	Event    model.Event     `json:"event"`
	NodeID   string          `json:"nodeId"`
	Strategy *model.Strategy `json:"strategy"`
}

func (o *Orchestrator) HandleEvent(event model.Event) error {
	// 执行策略
	startNode := o.Strategy.GetStartNode()
	if startNode == nil {
		return fmt.Errorf("strategy has no start node")
	}

	return o.executeNode(startNode, event)
}

func (o *Orchestrator) executeNode(node *model.StrategyNode, event model.Event) error {
	switch node.Type {
	case string(model.NodeTypeCondition):
		condType, ok := node.Config["type"].(string)
		if !ok {
			return fmt.Errorf("condition type not found in config")
		}

		cond, ok := condition.GetCondition(condType)
		if !ok {
			return fmt.Errorf("condition %s not registered", condType)
		}

		if cond.Evaluate(event, event.Context) {
			for _, nextID := range node.Next {
				nextNode := o.findNode(nextID)
				if nextNode == nil {
					return fmt.Errorf("next node %s not found", nextID)
				}
				if err := o.executeNode(nextNode, event); err != nil {
					return err
				}
			}
		}

	case string(model.NodeTypeAction):
		actionType, ok := node.Config["type"].(string)
		if !ok {
			return fmt.Errorf("action type not found in config")
		}

		actionHandler, ok := action.GetAction(actionType)
		if !ok {
			return fmt.Errorf("action %s not registered", actionType)
		}

		if err := actionHandler.Execute(event, node.Config); err != nil {
			return fmt.Errorf("failed to execute action: %v", err)
		}

	case string(model.NodeTypeDelay):
		duration, ok := node.Config["duration"].(string)
		if !ok {
			return fmt.Errorf("delay duration not found in config")
		}

		_, err := time.ParseDuration(duration)
		if err != nil {
			return fmt.Errorf("invalid delay duration: %v", err)
		}

		resumeMsg := ResumeMessage{
			Event:    event,
			NodeID:   node.Next[0],
			Strategy: o.Strategy,
		}

		msgBytes, err := json.Marshal(resumeMsg)
		if err != nil {
			return fmt.Errorf("failed to marshal resume message: %v", err)
		}

		// 发送延迟消息
		if err := o.Producer.Produce(o.DelayTopic, nil, msgBytes); err != nil {
			return fmt.Errorf("failed to produce delay message: %v", err)
		}

	case string(model.NodeTypeEventTrigger):
		eventType, ok := node.Config["eventType"].(string)
		if !ok {
			return fmt.Errorf("event type not found in config")
		}

		if eventType == event.Type {
			for _, nextID := range node.Next {
				nextNode := o.findNode(nextID)
				if nextNode == nil {
					return fmt.Errorf("next node %s not found", nextID)
				}
				if err := o.executeNode(nextNode, event); err != nil {
					return err
				}
			}
		}

	default:
		return fmt.Errorf("unknown node type: %s", node.Type)
	}

	return nil
}

func (o *Orchestrator) findNode(id string) *model.StrategyNode {
	for _, node := range o.Strategy.Nodes {
		if node.ID == id {
			return &node
		}
	}
	return nil
}

// ProcessResumeMessage processes a resume message from the delay queue
func (o *Orchestrator) ProcessResumeMessage(data []byte) error {
	var msg ResumeMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal resume message: %v", err)
	}

	node := o.findNode(msg.NodeID)
	if node == nil {
		return fmt.Errorf("resume node %s not found", msg.NodeID)
	}

	return o.executeNode(node, msg.Event)
}

func (o *Orchestrator) StartDelayScheduler(consumer ports.Consumer) {
	go func() {
		for {
			msg, err := consumer.Consume()
			if err != nil {
				utils.Logger.Printf("Error consuming delay message: %v", err)
				continue
			}
			if msg == nil {
				continue
			}

			if err := o.ProcessResumeMessage(msg.Value); err != nil {
				utils.Logger.Printf("Error handling resume message: %v", err)
			}
		}
	}()
}
