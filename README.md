### Proxy/chache for apt, npm, maven/grandl

example config  Docker_k8s/example.yaml

Build program
```
cp Docker_k8s/example yaam2.yaml
go mod tidy
CGO_ENABLED=0 go build -o yaam2 main.go
```