package main

import (
	"fmt"
	"os"
)

type Produser struct{
	file *os.File
}

func NewProducer(filename string) (*Produser, error){
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil{
		return nil, err
	}

	return &Produser{file: file}, nil
}

func (p *Produser) Close() error{
	return p.file.Close()
}

func (p *Produser) WriteFile(metricsStorage *MemStorage) error{

	for name, value := range metricsStorage.Counters{
		data := fmt.Sprintf("%s: %d\n", name, value)

		err := os.WriteFile(p.file.Name(), []byte(data), 0666)
		if err != nil{
			return err
		}
	}

	for name, value := range metricsStorage.Gauges{
		data:= fmt.Sprintf("%s: %f\n", name, value)
		err := os.WriteFile(p.file.Name(), []byte(data), 0666)
		if err != nil{
			return err
		}
	}

	return nil
}
