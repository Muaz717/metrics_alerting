package storage

import (
	"errors"
	"fmt"

	"github.com/Muaz717/metrics_alerting/internal/config"
)

type Metrics struct{
	ID		string		`json:"id"`
	MType 	string		`json:"type"`
	Delta	*int64		`json:"delta,omitempty"`
	Value	*float64	`json:"value,omitempty"`
}

// Интерфейс для взаимодействия с хранилищем
type Storage interface {
	SaveJSON(metrics Metrics) (*Metrics, error)
	GetJSON(metrics Metrics) (*Metrics, error)
	SaveMetric(mType, mName, mValue string) error
	GetMetric(mType, mName string) (string, error)
	GetAllMetrics() string
}

func NewStore(cfg *config.ServerCfg) (Storage, error){
	switch{
	case cfg.StorageCfg.FileStoragePath != "":
		store, err := NewFileStorage(cfg)
		if err != nil{
			return nil, fmt.Errorf("error creating file storage: %w", err)
		}
		return store, nil
	case cfg.StorageCfg.FileStoragePath == "":
		store, err := NewMemStorage(cfg)
		if err != nil{
			return nil, fmt.Errorf("error creating memory storage: %w", err)
		}
		return store, nil
	default:
		return nil, errors.New("error creating storage")
	}
}
