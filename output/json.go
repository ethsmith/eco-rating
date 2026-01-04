package output

import (
	"encoding/json"
	"os"

	"eco-rating/model"
)

func Export(players map[uint64]*model.PlayerStats, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(players)
}
