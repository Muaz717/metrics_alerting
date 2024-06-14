package storage

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Muaz717/metrics_alerting/internal/config"
	"github.com/Muaz717/metrics_alerting/internal/logger"
)

type FileStorage struct{
	MemStore *MemStorage
	savePath string
}

func NewFileStorage(cfg *config.ServerCfg) (*FileStorage, error){
	memStore, err := NewMemStorage(cfg)
	if err != nil{
		return nil, fmt.Errorf("error creating memory storage: %w", err)
	}

	fileStorage := &FileStorage{
		MemStore: memStore,
		savePath: cfg.StorageCfg.FileStoragePath,
	}

	if cfg.StorageCfg.Restore{
		err := LoadMetrics(fileStorage, cfg.StorageCfg.FileStoragePath)
		if err != nil{
			return nil, fmt.Errorf("error loading metrics: %w", err)
		}
	}
	return fileStorage, nil
}

func LoadMetrics(f *FileStorage, filePath string) error{
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			_, err := os.Create(filePath)
			if err != nil {
				return fmt.Errorf("error creating storage file: %w", err)
			}
			return nil
		} else {
			return fmt.Errorf("unexpected error while creating storage file: %w", err)
		}
	} else {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("cannot read storage file: %w", err)
		}

		if err := json.Unmarshal(data, f); err != nil {
			return fmt.Errorf("cannot unmarshal storage file, file is probably empty: %w", err)
		}
		return nil
	}
}

func SaveMetrics(s Storage, filePath string) error{
	err := logger.Initialize("info")
	if err != nil{
		return fmt.Errorf("failed to init logger: %w", err)
	}

	data, err := json.MarshalIndent(s, "", " ")
	if err != nil{
		logger.Log.Info("cannot marshal storage")
		return fmt.Errorf("cannot marshal storage: %w", err)
	}

	err = os.WriteFile(filePath, data, 0666)
	if err != nil {
		logger.Log.Info("cannot save storage to file")
		return fmt.Errorf("cannot save storage to file: %w", err)
	}

	return nil
}

func (f *FileStorage) SaveMetric(mType, mName, mValue string) error {
	if err := f.MemStore.SaveMetric(mType, mName, mValue); err != nil{
		return fmt.Errorf("cannot save metric: %w", err)
	}
	return nil
}

func (f *FileStorage) GetMetric(mType, mName string) (string, error) {
	val, err := f.MemStore.GetMetric(mType, mName)
	if err != nil {
		return "", fmt.Errorf("cannot get metric: %w", err)
	}
	return val, nil
}

func (f *FileStorage) GetAllMetrics() string {
	return f.MemStore.GetAllMetrics()
}

func (f *FileStorage) SaveJson(metrics Metrics) (*Metrics, error){
	response, err := f.MemStore.SaveJson(metrics)
	if err != nil {
		return nil, fmt.Errorf("cannot dave json metric: %w", err)
	}
	return response, nil
}

func (f *FileStorage) GetJson(metrics Metrics) (*Metrics, error){
	response, err := f.MemStore.GetJson(metrics)
	if err != nil {
		return nil, fmt.Errorf("cannot dave json metric: %w", err)
	}
	return response, nil
}
