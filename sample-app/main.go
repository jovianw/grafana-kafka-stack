package main

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	otellog "go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.23.1"
)

const (
	otlpEndpoint = "localhost:4317"
)

func main() {
	// Some code taken from https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp#New:~:text=by%20WithTLSClientConfig%20option.-,Example%20%C2%B6,-Index%20%C2%B6
	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("customer"),
			semconv.ServiceVersion("v1.0.0"),
		),
	)
	if err != nil {
		log.Fatalf("failed to create resource: %v", err)
	}

	// Logs
	logExporter, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpoint(otlpEndpoint),
		otlploghttp.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("failed to create log exporter: %v", err)
	}

	logProcessor := sdklog.NewBatchProcessor(logExporter)
	logProvider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(logProcessor),
		sdklog.WithResource(res),
	)

	logger := logProvider.Logger("sample-logger")

	// Metrics
	metricExporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint(otlpEndpoint),
		otlpmetrichttp.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("failed to create metric exporter: %v", err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter)),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)

	meter := meterProvider.Meter("sample-meter")

	counter, err := meter.Int64Counter("sample_counter")
	if err != nil {
		log.Fatalf("failed to create counter: %v", err)
	}

	// Emit
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		// Emit a log
		record := otellog.Record{}
		record.SetBody(otellog.StringValue("Hello, this is a sample log message!"))
		record.SetSeverity(otellog.SeverityInfo)
		record.SetEventName("sample-log-event")
		record.SetTimestamp(time.Now())
		logger.Emit(
			ctx,
			record,
		)

		// Emit a metric
		counter.Add(ctx, 1)

		// Simulate some work
		time.Sleep(500 * time.Millisecond)
	}
}
