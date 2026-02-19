package main

import (
	"context"
	"net/http"

	"github.com/OscarVillanueva/goapi/internal/app/handlers"
	"github.com/OscarVillanueva/goapi/internal/platform"

	"go.opentelemetry.io/otel"
	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
	"github.com/riandyrn/otelchi"
	otelchimetric "github.com/riandyrn/otelchi/metric"
)

func main()  {
	serverName := "ecommerce-backend"
	ctx := context.Background()
	log.SetReportCaller(true)

	tp, err := platform.InitTracer(ctx)
	if err != nil {
		log.Fatal(err)
		return
	}

	mp := platform.InitMeter()
	otel.SetMeterProvider(mp)

	baseCfg := otelchimetric.NewBaseConfig(serverName, otelchimetric.WithMeterProvider(mp))

	var r *chi.Mux = chi.NewRouter()
	r.Use(
		otelchi.Middleware(serverName, otelchi.WithChiRoutes(r)),
		otelchimetric.NewRequestDurationMillis(baseCfg),
		otelchimetric.NewRequestInFlight(baseCfg),
		otelchimetric.NewResponseSizeBytes(baseCfg),
	)

	handlers.Router(r)

	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Error("Error shutting down tracer: %v", err)
		}
	}()

	log.Info("Starting GO API Server")

	if err := http.ListenAndServe("backend:4321", r); err != nil {
		log.Error(err)
	}
}


