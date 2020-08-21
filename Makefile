VERSION=`cat VERSION`
up:
	kind create cluster
down:
	kind delete cluster
jaeger-install:
	kubectl create ns tracing || true
	helm repo add jaegertracing https://jaegertracing.github.io/helm-charts
	helm install jaeger jaegertracing/jaeger -n tracing \
	--set cassandra.config.max_heap_size=1024M \
  	--set cassandra.config.heap_new_size=256M \
  	--set cassandra.resources.requests.memory=2048Mi \
  	--set cassandra.resources.requests.cpu=0.4 \
  	--set cassandra.resources.limits.memory=2048Mi \
  	--set cassandra.resources.limits.cpu=0.4 \
	--set provisionDataStore.kafka=true \
  	--set ingester.enabled=true
jaeger-bomb-install:
	cd helm && helm install jaeger-bomb . && cd ../
docker-build:
	docker build . -t tibbar/jaeger-bomb:$(VERSION)
docker-push:
	docker push tibbar/jaeger-bomb:$(VERSION)
publish: docker-build docker-push
