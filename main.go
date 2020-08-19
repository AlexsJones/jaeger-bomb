package main

import (
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
	"math/rand"
	"os"
	"time"
)

var (
	jLogger = jaegerlog.StdLogger
	jMetricsFactory = metrics.NullFactory

)
func emitTrace() {
	tracer := opentracing.GlobalTracer()
	span := tracer.StartSpan("jaeger-bomb-parent-trace")
	defer span.Finish()
	childCount := rand.Intn(15) // n will be between 0 and 5

	lastParent := span.Context()
	for i :=0; i < childCount; i++ {
		// Create a Child Span. Note that we're using the ChildOf option.
		childSpan := tracer.StartSpan(
			fmt.Sprintf("child-%d",i),
			opentracing.ChildOf(lastParent),
		)
		sleepTime := rand.Intn(5000)
		// Delay in the child spans
		time.Sleep(time.Duration(sleepTime) * time.Millisecond)

		defer childSpan.Finish()
		lastParent = childSpan.Context()
	}
	jLogger.Infof("Generated %d child spans",childCount)
}
func main() {

	cfg := jaegercfg.Configuration{
		ServiceName: "jaeger-bomb",
		Sampler:     &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeRemote,
			Param: 1,
		},
		Reporter:    &jaegercfg.ReporterConfig{
			LogSpans: true,
		},
	}

	tracer, closer, err := cfg.NewTracer(
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
	if err != nil {
		os.Exit(1)
	}
	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()

	rand.Seed(time.Now().UnixNano())

	for {

		emitTrace()
		n := rand.Intn(5) // n will be between 0 and 5
		time.Sleep(time.Duration(n) * time.Second)
	}
}