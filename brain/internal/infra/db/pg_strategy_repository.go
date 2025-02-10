package db

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/sjtutwilight/Twilight/brain/internal/domain/model"
)

type PgStrategyRepository struct {
	db *sql.DB
}

func NewPgStrategyRepository(db *sql.DB) *PgStrategyRepository {
	return &PgStrategyRepository{db: db}
}

func (r *PgStrategyRepository) GetStrategyByID(id string) (*model.Strategy, error) {
	var strategy model.Strategy
	var configJSON []byte

	err := r.db.QueryRow(`
		SELECT id, name, description, config
		FROM strategies
		WHERE id = $1 AND active = true`,
		id,
	).Scan(&strategy.ID, &strategy.Name, &strategy.Description, &configJSON)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query strategy: %v", err)
	}

	var config struct {
		StartNode *model.StrategyNode  `json:"startNode"`
		Nodes     []model.StrategyNode `json:"nodes"`
	}
	if err := json.Unmarshal(configJSON, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	strategy.StartNode = config.StartNode
	strategy.Nodes = config.Nodes

	return &strategy, nil
}

func (r *PgStrategyRepository) CreateStrategy(strategy *model.Strategy) error {
	config := struct {
		StartNode *model.StrategyNode  `json:"startNode"`
		Nodes     []model.StrategyNode `json:"nodes"`
	}{
		StartNode: strategy.StartNode,
		Nodes:     strategy.Nodes,
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	err = r.db.QueryRow(`
		INSERT INTO strategies (name, description, config)
		VALUES ($1, $2, $3)
		RETURNING id`,
		strategy.Name, strategy.Description, configJSON,
	).Scan(&strategy.ID)

	if err != nil {
		return fmt.Errorf("failed to insert strategy: %v", err)
	}

	return nil
}

func (r *PgStrategyRepository) ListStrategies() ([]*model.Strategy, error) {
	rows, err := r.db.Query(`
		SELECT id, name, description, config
		FROM strategies
		WHERE active = true
		ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query strategies: %v", err)
	}
	defer rows.Close()

	var strategies []*model.Strategy
	for rows.Next() {
		var strategy model.Strategy
		var configJSON []byte

		err := rows.Scan(&strategy.ID, &strategy.Name, &strategy.Description, &configJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan strategy: %v", err)
		}

		var config struct {
			StartNode *model.StrategyNode  `json:"startNode"`
			Nodes     []model.StrategyNode `json:"nodes"`
		}
		if err := json.Unmarshal(configJSON, &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %v", err)
		}

		strategy.StartNode = config.StartNode
		strategy.Nodes = config.Nodes
		strategies = append(strategies, &strategy)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating strategies: %v", err)
	}

	return strategies, nil
}

// GetAllStrategies 获取所有策略
func (r *PgStrategyRepository) GetAllStrategies() ([]*model.Strategy, error) {
	rows, err := r.db.Query("SELECT id, name, description, config FROM strategies")
	if err != nil {
		return nil, fmt.Errorf("query strategies failed: %v", err)
	}
	defer rows.Close()

	var strategies []*model.Strategy
	for rows.Next() {
		var strategy model.Strategy
		var configStr string
		err := rows.Scan(&strategy.ID, &strategy.Name, &strategy.Description, &configStr)
		if err != nil {
			return nil, fmt.Errorf("scan strategy failed: %v", err)
		}

		err = json.Unmarshal([]byte(configStr), &strategy)
		if err != nil {
			return nil, fmt.Errorf("unmarshal strategy config failed: %v", err)
		}

		strategies = append(strategies, &strategy)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate strategies failed: %v", err)
	}

	return strategies, nil
}
