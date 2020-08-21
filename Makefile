VERSION=`cat VERSION`
up:
	kind create cluster
down:
	kind delete cluster
jaeger-install:
	kubectl create ns tracing || true
	helm repo add jaegertracing https://jaegertracing.github.io/helm-charts
	helm install jaeger jaegertracing/jaeger -n tracing
jaeger-bomb-install:
	cd helm && helm install jaeger-bomb . && cd ../
docker-build:
	docker build . -t tibbar/jaeger-bomb:$(VERSION)
docker-push:
	docker push tibbar/jaeger-bomb:$(VERSION)
publish: docker-build docker-push
