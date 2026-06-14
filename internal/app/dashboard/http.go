package dashboard

import "net/http"

func WriteBootstrap(w http.ResponseWriter, opts BootstrapOptions) error {
	body, err := BootstrapHTML(opts)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(body)
	return nil
}
