package middleware

import (
	"github.com/ennismar/go-helper/pkg/tracing"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer(tracing.Middleware)
