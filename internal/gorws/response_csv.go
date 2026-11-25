package rws

import (
	"encoding/csv"
	"io"
	"strings"
)

type CsvResp struct {
	Data   [][]string
	Source string
}

func (resp *CsvResp) Decode(reader io.Reader) error {
	all, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	source := strings.TrimSuffix(string(all), "\nEOF")
	csvReader := csv.NewReader(strings.NewReader(source))
	readAll, err := csvReader.ReadAll()
	if err != nil {
		return err
	}
	resp.Data = readAll
	resp.Source = source
	return nil
}
