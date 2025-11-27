//go:build !webui

package main

import (
	"net/http"
)

// registerWebUI is an empty stub because `webui` tag has not been specified.
//
//nolint:unused
func registerWebUI(hdl http.Handler) (http.Handler, error) {
	return hdl, nil
}
