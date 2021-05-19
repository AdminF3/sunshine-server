package contract

import (
	"fmt"
	"os"
	"testing"

	"stageai.tech/sunshine/sunshine/models"
)

func TestMain(m *testing.M) {
	statusCode := m.Run()
	if err := models.ClearTestSchemas(); err != nil {
		fmt.Printf("Clear test schemas: %v", err)
		statusCode = 1
	}
	os.Exit(statusCode)
}
