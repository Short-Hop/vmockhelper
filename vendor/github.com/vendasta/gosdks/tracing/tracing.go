package tracing

import (
	"context"
	"log"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/vendasta/gosdks/config"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

// Setup is used in the serverconfig sdk when creating a server.
// It handles setting up trace exporters, samplers, and any labels that are to appear on all traces.
func Setup(appID string, projectID string) error {
	labels := &stackdriver.Labels{}
	labels.Set("appId", appID, "App ID")
	labels.Set("env", config.CurEnv().Name(), "Environment")
	labels.Set("podId", config.GetGkePodName(), "GKE Pod Id")
	labels.Set("namespace", config.GetGkeNamespace(), "GKE Namespace")
	labels.Set("tag", config.GetTag(), "GCR Tag")

	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID:    projectID,
		MetricPrefix: "Vendasta",
		DefaultTraceAttributes: map[string]interface{}{
			"appId":     appID,
			"env":       config.CurEnv().Name(),
			"podId":     config.GetGkePodName(),
			"namespace": config.GetGkeNamespace(),
			"tag":       config.GetTag(),
		},
		DefaultMonitoringLabels: labels,
	})
	if err != nil {
		log.Printf("Error creating stackdriver exporter %s", err.Error())
	} else {
		tracingLowerBound := 1
		tracingHigherBound := 5
		if config.IsProd() {
			tracingLowerBound = 5
			tracingHigherBound = 100
		}
		view.RegisterExporter(exporter)
		trace.RegisterExporter(exporter)
		trace.ApplyConfig(trace.Config{DefaultSampler: NewGuaranteedThroughputProbabilisticSampler(float64(tracingLowerBound), float64(tracingHigherBound), 0.01).Sampler})
	}
	return nil
}

func FromContext(ctx context.Context) *trace.Span {
	span := trace.FromContext(ctx)
	span.Annotate(nil, "Namespace: "+config.GetGkeNamespace())
	return span
}

func DisableLogTracing() (context.Context, func()) {
	ctx, span := trace.StartSpan(context.Background(), "this span will not be exported",
		trace.WithSampler(trace.NeverSample()))
	return ctx, span.End
}

func StartSpan(ctx context.Context, name string, opts ...trace.StartOption) (context.Context, *trace.Span) {
	c, span := trace.StartSpan(ctx, name, opts...)
	span.Annotate(nil, "Namespace: "+config.GetGkeNamespace())
	return c, span
}
