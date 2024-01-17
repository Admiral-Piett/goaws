package smoke_tests

import (
	"net/http/httptest"

	"github.com/Admiral-Piett/goaws/app/router"
)

func generateServer() *httptest.Server {
	return httptest.NewServer(router.New())
}
