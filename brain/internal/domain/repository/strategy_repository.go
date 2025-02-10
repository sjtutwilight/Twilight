package repository

import "github.com/sjtutwilight/Twilight/brain/internal/domain/model"

type StrategyRepository interface {
	GetStrategyByID(id string) (*model.Strategy, error)
	CreateStrategy(strategy *model.Strategy) error
	ListStrategies() ([]*model.Strategy, error)
	// 未来可添加List, Save等方法
}
