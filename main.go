package main

import (
	"flag"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
	"log"
	"math/rand"
	"net/http"
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

		url := "http://jaeger-bomb-server:8082/publish"
		req, _ := http.NewRequest("GET", url, nil)
		// Set some tags on the clientSpan to annotate that it's the client span. The additional HTTP tags are useful for debugging purposes.
		ext.SpanKindRPCClient.Set(childSpan)
		ext.HTTPUrl.Set(childSpan, url)
		ext.HTTPMethod.Set(childSpan, "GET")

		// Inject the client span context into the headers
		tracer.Inject(childSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))

		resp, _ := http.DefaultClient.Do(req)
		if resp.StatusCode != 200 {
			jLogger.Error(resp.Status)
		}
		defer childSpan.Finish()
		lastParent = childSpan.Context()
	}
	jLogger.Infof("Generated %d child spans",childCount)
}

func serverMode() {
	tracer := opentracing.GlobalTracer()
	http.HandleFunc("/publish", func(w http.ResponseWriter, r *http.Request) {

		spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		serverSpan := tracer.StartSpan("server", ext.RPCServerOption(spanCtx))
		defer serverSpan.Finish()
	})
	log.Fatal(http.ListenAndServe(":8082", nil))
}

func main() {

	isServer := flag.Bool("server",false,"Run the application in server mode for receiving spans")

	flag.Parse()

	var serviceName string

	if *isServer {
		serviceName = "jaeger-bomb-server"
	} else {
		serviceName = "jaeger-bomb"
	}

	cfg := jaegercfg.Configuration{
		ServiceName: serviceName,
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

	if *isServer {
		jLogger.Infof("Running in server mode")
		serverMode()
	}else {
		for {
			jLogger.Infof("Running in client mode")
			emitTrace()
			n := rand.Intn(5) // n will be between 0 and 5
			time.Sleep(time.Duration(n) * time.Second)
		}
	}
}