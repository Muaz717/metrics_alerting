package storage

import (
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/Muaz717/metrics_alerting/internal/config"
	"github.com/Muaz717/metrics_alerting/internal/logger"

	"go.uber.org/zap"
)

// Тип хранилища для метрик
type MemStorage struct {
	mx sync.Mutex
	Gauges map[string]float64
	Counters map[string]int64
}

func NewMemStorage(cfg *config.ServerCfg) (*MemStorage, error){
	return &MemStorage{
		mx: sync.Mutex{},
		Gauges: make(map[string]float64),
		Counters: make(map[string]int64),
	}, nil
}

func (s *MemStorage) SaveMetric(mType string, mName string, mValue string) error{
	switch mType{
	case "gauge":
		if err := s.UpdateGauge(mName, mValue); err != nil{
			return fmt.Errorf("failed to save gauge: %w", err)
		}
	case "counter":
		if err := s.UpdateCounter(mName, mValue); err != nil{
			return fmt.Errorf("failed to save counter: %w", err)
		}
	default:
		return fmt.Errorf("wrong metric type: %v", mType)
	}
	return nil
}

func (s *MemStorage) UpdateGauge(name string, value string) error{
	valueFloat, err := strconv.ParseFloat(value, 64)
	if err != nil{
		return fmt.Errorf("got error parsing int64 value for counter metric: %w", err)
	}

	s.mx.Lock()
	s.Gauges[name] = valueFloat
	s.mx.Unlock()

	return nil
}

func (s *MemStorage) UpdateCounter(name string, value string) error{
	valueInt64, err := strconv.ParseInt(value, 10, 64)
	if err != nil{
		return fmt.Errorf("got error parsing int64 value for counter metric: %w", err)
	}

	s.mx.Lock()
	s.Counters[name] += valueInt64
	s.mx.Unlock()

	return nil
}

func (s *MemStorage) GetMetric(mType, mName string) (string, error) {
	var html string
	var ErrItemNotFound = errors.New("item not found")

	switch mType {
	case "gauge":
		if mValue, ok := s.Gauges[mName]; !ok {
			return "", ErrItemNotFound
		} else {
			html = strconv.FormatFloat(mValue, 'f', -1, 64)
			return html, nil
		}
	case "counter":
		if mValue, ok := s.Counters[mName]; !ok {
			return "", ErrItemNotFound
		} else {
			html = strconv.FormatInt(mValue, 10)
			return html, nil
		}
	default:
		return "", nil
	}
}

func (s *MemStorage) GetAllMetrics() string{
	html := "<h3>Gauge:</h3>"
	for mName, mValue := range s.Gauges {
		html += (mName + ":" + strconv.FormatFloat(mValue, 'f', -1, 64) + "<br>")
	}
	html += "<h3>Counter:</h3>"
	for mName, mValue := range s.Counters {
		html += (mName + ":" + strconv.FormatInt(mValue, 10) + "<br>")
	}
	return html
}

func (s *MemStorage) SaveJSON(metrics Metrics) (*Metrics, error){
	switch metrics.MType{
	case "gauge":
		resp, err := s.UpdateGaugeJSON(metrics)
		if err != nil{
			return nil, fmt.Errorf("error save gauge json metric: %w", err)
		}
		return resp, nil
	case "counter":
		resp, err := s.UpdateCounterJSON(metrics)
		if err != nil{
			return nil, fmt.Errorf("error save gauge json metric: %w", err)
		}
		return resp, nil
	default:
		logger.Log.Info("Wrong metric type")
		return nil, fmt.Errorf("wrong metric type: %v", metrics.MType)
	}
}

func (s *MemStorage) UpdateGaugeJSON(metrics Metrics) (*Metrics, error){
	s.mx.Lock()
	defer s.mx.Unlock()

	response := Metrics{
		ID: metrics.ID,
		MType: metrics.MType,
		Value: metrics.Value,
	}

	if response.ID == ""{
		logger.Log.Info("Forgot metric name")
		return nil, fmt.Errorf("forgot metric name: %v", response.ID)
	}
	if response.Value == nil{
		return nil, fmt.Errorf("no value: %v", response.Value)
	}

	s.Gauges[response.ID] = *metrics.Value

	return &response, nil

}

func (s *MemStorage) UpdateCounterJSON(metrics Metrics) (*Metrics, error){
	s.mx.Lock()
	defer s.mx.Unlock()

	response := Metrics{
		ID: metrics.ID,
		MType: metrics.MType,
		Delta: metrics.Delta,
	}

	if response.ID == ""{
		logger.Log.Info("Forgot metric name")
		return nil, fmt.Errorf("forgot metric name: %v", response.ID)
	}

	if response.Delta == nil{
		return nil, fmt.Errorf("no value: %v", response.Value)
	}
	if value, ok := s.Counters[response.ID]; ok{
		newValue := *response.Delta + value
		response.Delta = &newValue
	}
	s.Counters[response.ID] = *response.Delta

	return &response, nil

}

func (s *MemStorage) GetJSON(metrics Metrics) (*Metrics, error){
	switch metrics.MType{
	case "gauge":
		resp, err := s.ValueGaugeJSON(metrics)
		if err != nil{
			return nil, fmt.Errorf("error save gauge json metric: %w", err)
		}
		return resp, nil
	case "counter":
		resp, err := s.ValueCounterJSON(metrics)
		if err != nil{
			return nil, fmt.Errorf("error save gauge json metric: %w", err)
		}
		return resp, nil
	default:
		logger.Log.Info("Wrong metric type")
		return nil, fmt.Errorf("wrong metric type: %v", metrics.MType)
	}
}

func (s *MemStorage) ValueGaugeJSON(metrics Metrics) (*Metrics, error){
	s.mx.Lock()
	defer s.mx.Unlock()

	response := Metrics{
		ID: metrics.ID,
		MType: metrics.MType,
	}

	if _, ok := s.Gauges[response.ID]; !ok{
		logger.Log.Info("No gauge metric with this id", zap.String("metricName", response.ID) )
		return nil, fmt.Errorf("no gauge metric with this id : %v", response.ID)
	}
	value := s.Gauges[response.ID]
	response.Value = &value

	return &response, nil

}

func (s *MemStorage) ValueCounterJSON(metrics Metrics) (*Metrics, error){
	s.mx.Lock()
	defer s.mx.Unlock()

	response := Metrics{
		ID: metrics.ID,
		MType: metrics.MType,
	}

	if _, ok := s.Counters[response.ID]; !ok{
		logger.Log.Info("No counter metric with this id", zap.String("metricName", response.ID) )
		return nil, fmt.Errorf("no counter metric with this id: %v", response.ID)
	}
	delta := s.Counters[response.ID]
	response.Delta = &delta

	return &response, nil

}
