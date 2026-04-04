package telemetry

import (
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Telemetry is the "Observability Bundle" for the entire system.
type Telemetry struct {
	Logger  *zap.Logger
	Metrics *Metrics
	Tracer  trace.Tracer
	Meter   metric.Meter
}

// ScopedLogger returns a logger with the module name already baked in.
func (t Telemetry) ScopedLogger(moduleName string) *zap.Logger {
	return t.Logger.With(zap.String("module", moduleName))
}
