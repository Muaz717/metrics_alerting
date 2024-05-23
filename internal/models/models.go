package models
// Мапа для хранения имеющегося в структуре counter значения
var m = map[string]int64{}

type Metrics struct{
	ID		string		`json:"id"`
	MType 	string		`json:"type"`
	Delta	*int64		`json:"delta,omitempty"`
	Value	*float64	`json:"value,omitempty"`
}
