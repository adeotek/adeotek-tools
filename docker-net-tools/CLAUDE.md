# CLAUDE.md — docker-net-tools

A Debian-slim Docker image pre-loaded with network diagnostic and troubleshooting tools (tcpdump, tshark, nmap, iperf3, mtr, etc.). No application code — changes are made to the `Dockerfile`.

See [README.md](README.md) for the full tool inventory and usage examples.

## Build Commands

```bash
# Build image (tag defaults to docker-registry/docker-net-tools:<version>)
./build.sh v1.0.0

# Build and push to registry
./build.sh v1.0.0 --push

# Manual build
docker build -t docker-net-tools:dev .
```

## Run

```bash
# Interactive shell
docker run -it --rm docker-net-tools:dev

# With host network access (required for most diagnostic tools)
docker run -it --rm --network host docker-net-tools:dev
```

## Key Notes / Gotchas

- **Root user**: container runs as root; `chmod -R g=u /root` is applied for OpenShift compatibility
- **tshark permissions**: `dumpcap` is chowned to root:root so packet capture works inside the container without `--privileged` for most cases; some environments still need `--cap-add NET_RAW`
- No tests — validate changes by building and running the container interactively
