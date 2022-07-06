# Http Server - Proxy Replica

Replicate the same request over main proxy and replicas


# Example
proxies.yaml
```yaml
server:
  # Listen to 0.0.0.0:80 
  bind_address: ":http"

  # For Let's Encrypt (automatic)
  # When enabled, exposes :https (443 port) too
  autotls:
    enabled: true
    domains: [ 'host1.your-domain.tld', 'host2.your-domain.tld' ]
    email: "email@your-domain.tld"

# Main Proxy (Required)
# Wait for response (synchronous)
main:
  url: http://localhost:8080
  verify_ssl: false
  timeout: 10s

# Secondary Proxies (Optional)
# Asynchronous (no wait)
replicas:
  - url: http://localhost:8081
    verify_ssl: false
    timeout: 5s

  - url: http://localhost:8082
    verify_ssl: false
    timeout: 5s
```

Build from source
```bash
# any platform (Windows, macOS or Linux)
make go

# for linux only
make go_linux

# docker image
make docker_image
```

Run from build
```bash
# macOS / Linux
./go-proxy-replica -config ./proxies.yaml

# Windows
./go-proxy-replica.exe -config ./proxies.yaml

```

Run from docker
```bash
docker run -ti \
        --name go-proxy-replica \
        --privileged \
        -p 80:80 \
        -v ./proxies.yaml:/app/proxies.yaml:ro \
        simonops/go-proxy-replica:latest

```
