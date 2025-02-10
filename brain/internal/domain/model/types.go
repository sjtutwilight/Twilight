package model

type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp int64                  `json:"timestamp"`
	Context   map[string]interface{} `json:"context,omitempty"` // 用于存储事件处理过程中的上下文数据
}

type Strategy struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	StartNode   *StrategyNode  `json:"startNode"`
	Nodes       []StrategyNode `json:"nodes"`
}

func (s *Strategy) GetStartNode() *StrategyNode {
	return s.StartNode
}

type StrategyNode struct {
	ID     string                 `json:"id"`
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
	Next   []string               `json:"next"`
}

type NodeType string

const (
	NodeTypeCondition    NodeType = "condition"
	NodeTypeAction       NodeType = "action"
	NodeTypeDelay        NodeType = "delay"
	NodeTypeEventTrigger NodeType = "eventTrigger"
)
