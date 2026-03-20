package httpapi

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
)

func NewAPI(mux *http.ServeMux) huma.API {
	config := huma.DefaultConfig("AI Chatbot Platform API", "0.1.0")
	config.Info.Description = "Single-tenant chatbot administration and chat runtime API."
	config.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearer": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
		},
	}

	api := humago.New(mux, config)
	api.UseMiddleware(RequestIDMiddleware)
	api.UseMiddleware(OtelTracingMiddleware)
	return api
}
