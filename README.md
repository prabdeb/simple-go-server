# simple-go-server

This is a simple GO server to display basic OS information, useful for testing deployment setup with less spent in Memory/CPU

## Docker Execution

```sh
docker run --rm -it -p 8080:8080 prabdeb/simple-go-server:latest
```

## Helm Installation

```sh
cd helm/simple-go-server
helm install --name sample-app .
```

## Kustomize Installation

```sh
kubectl apply -k kustomize/
```

**Note:** Kustomize is available on kubectl v1.14