package app

import (
	"encoding/json"
	"io"
	"net/http"
)

func (api *runtimeHTTPAPI) handleMihomoNativeConfig(w http.ResponseWriter, req *http.Request) {
	runtime := api.service.settings.runtime
	switch req.Method {
	case http.MethodGet:
		settings, err := mihomoNativeSettings(runtime)
		if err != nil {
			writeHTTPError(w, err, http.StatusInternalServerError)
			return
		}
		writeJSON(w, settings)
	case http.MethodPost, http.MethodPut:
		var settings mihomoNativeSettingsView
		if !readJSON(w, req, &settings) {
			return
		}
		next, err := updateMihomoNativeSettings(req.Context(), runtime, settings)
		if err != nil {
			writeHTTPError(w, err, http.StatusBadRequest)
			return
		}
		writeJSON(w, next)
	default:
		methodNotAllowed(w, http.MethodGet+", "+http.MethodPost+", "+http.MethodPut)
	}
}

func readJSON(w http.ResponseWriter, req *http.Request, dst any) bool {
	defer req.Body.Close()
	body, err := io.ReadAll(io.LimitReader(req.Body, 1<<20))
	if err != nil {
		writeHTTPError(w, invalidArgument("read request body", err), http.StatusBadRequest)
		return false
	}
	if len(body) == 0 {
		writeHTTPError(w, invalidArgument("request body is required", nil), http.StatusBadRequest)
		return false
	}
	if err := json.Unmarshal(body, dst); err != nil {
		writeHTTPError(w, invalidArgument("parse request body", err), http.StatusBadRequest)
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		writeHTTPError(w, internalError("marshal response", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}
