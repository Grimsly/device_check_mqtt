package loader

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"

	"device_check_mqtt/internal/store"
)

// Reads a CSV file and registers each device into the storage store
func LoadDevicesFromCSV(path string, s store.DeviceStorageStore) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("Failed to open %s: %w", path, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return 0, fmt.Errorf("Failed to parse CSV: %w", err)
	}

	// If the device records has less than 2 lines,
	// it means that either file is empty of there's only the header
	if len(records) < 2 {
		return 0, fmt.Errorf("CSV has no data rows")
	}

	// Check if the header is device_id. If it's not, there's a chance that this CSV file is wrong.
	if records[0][0] != "device_id" {
		return 0, fmt.Errorf("CSV does not contain device IDs")
	}

	loaded := 0
	// Skip header
	for index := 1; index < len(records); index++ {
		row := records[index]
		if len(row) == 0 || row[0] == "" {
			slog.Warn("Skipping empty row", "line", index+1)
			continue
		}
		if err := s.AddDevice(row[0]); err != nil {
			slog.Warn("Skipping device", "device_id", row[0], "error", err)
			continue
		}
		loaded++
	}

	return loaded, nil
}
