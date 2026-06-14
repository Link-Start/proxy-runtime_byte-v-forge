package app

import (
	"net/http"

	dashboardapp "github.com/byte-v-forge/proxy-runtime/internal/app/dashboard"
)

func applyMihomoControllerAuth(req *http.Request, token string) {
	dashboardapp.ApplyControllerAuthorization(req, token)
}
