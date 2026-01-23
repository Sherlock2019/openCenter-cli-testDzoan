# Mason Container Imageary tools to deploy a Kubernetes cluster.

## Build

### Podman

```bash
cd hack/mason
podman build -t mason:latest .
```

### Docker

```bash
cd hack/mason
docker build -t mason:latest .
```

## Run

```bash
# Podman
podman run -it --rm -v $(pwd):/deploy mason:latest

# Docker
docker run -it --rm -v $(pwd):/deploy mason:latest
```


Container image with all necess