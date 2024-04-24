package smoke_tests

import (
	"testing"

	"github.com/Admiral-Piett/goaws/app/utils"
)

func TestMain(m *testing.M) {
	utils.InitializeDecoders()
	m.Run()
}
