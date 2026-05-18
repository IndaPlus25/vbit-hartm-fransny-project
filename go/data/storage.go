package data

import (
	"io"
	"os"
	"sort"
	"trading-bot/types"

	"github.com/parquet-go/parquet-go"
)

func LoadData(filename string) ([]types.Bar, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := parquet.NewGenericReader[types.Bar](file)
	defer reader.Close()

	var allBars []types.Bar
	buf := make([]types.Bar, 1000)

	for {
		n, err := reader.Read(buf)
		allBars = append(allBars, buf[:n]...)

		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	//sorts allBars with pdqsort to make sure order is correct in parquet file
	sort.Slice(allBars, func(i, j int) bool {
		return allBars[i].Timestamp < allBars[j].Timestamp
	})
	return allBars, nil
}
