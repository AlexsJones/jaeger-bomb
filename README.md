# jaeger-bomb

A simple app to create Jaeger spans with [jaeger-client-go](https://github.com/jaegertracing/jaeger-client-go) in your kubernetes cluster to test your jaeger install is all hunky-dory.

This app will create a nest of spans every few seconds and send them to a server running on another pod.

```
cd helm
helm install jaeger-bomb . --set=jaeger.agent.connectionstring="mycollector.svc:9999"
```

## How to use this repository 💅

- This repository contains the golang code for creating the spans, the Dockerfile and the helm chart to get it into kubernetes.
- You may find you want to tweak the helm configuration to get it to work for your infrastructure.

### Set this example up in an existing cluster

#### Requirements
- helm
- kubectl

```
cd helm
helm install jaeger-bomb . --set=jaeger.agent.connectionstring="mycollector.svc:9999"
```

### Setup the toy example locally 🚀

This will setup a tiny Jaeger production like instance locally using the Jaeger helm chart.

#### Requirements
- helm
- docker
- kind
- kubectl



### How it works 👩🏻‍💻

Guts of the code... spits out a bunch of spans over and over.

 
```go
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
```

![](images/3.png)

- `make up`
- `make jaeger-install`
- `make jaeger-bomb-install`

_At this point you'll have a KIND cluster with a tracing namespace full of Jaeger components. Our helm chart jaeger-bomb will be sending spans from the
default namespace into the jaeger-collector in the tracing namespace_


![](images/1.png)

![](images/2.png)
