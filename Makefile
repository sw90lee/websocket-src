# ImageName
img = sw90lee/websocket-go
tag = 0.9.1


hello:
	echo ${img}

build:
	go build -o bin/main main.go

run:
	go run main.go

image:
	docker build -t ${img}:${tag} .
	docker push ${img}:${tag}
	helm upgrade websocket -f .\helm\websocket-go\values.yaml  ./helm/websocket-go-chart-0.1.0.tgz
	kubectl get pod

#pod:
#	helm upgrade websocket -f .\helm\websocket-go\values.yaml  ./helm/websocket-go-chart-0.1.0.tgz
#	kubectl get pod


cc:
	echo "Compiling for every OS and Platform"
	set GOOS=linux GOARCH=arm go build -o bin/main-linux-arm main.go
	set GOOS=linux GOARCH=arm64 go build -o bin/main-linux-arm64 main.go
	set GOOS=freebsd GOARCH=386 go build -o bin/main-freebsd-386 main.go



cclinux:
	echo "Compiling for every OS and Platform"
	GOOS=linux GOARCH=arm go build -o bin/main-linux-arm main.go
	GOOS=linux GOARCH=arm64 go build -o bin/main-linux-arm64 main.go
	GOOS=freebsd GOARCH=386 go build -o bin/main-freebsd-386 main.go


all: hello build