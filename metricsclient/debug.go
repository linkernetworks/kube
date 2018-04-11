package metricsclient

import (
	"encoding/json"
	"os"
)

// PrintPretty prints v as formatted JSON
func PrintPretty(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	os.Stdout.Write(data)
	return nil
}
