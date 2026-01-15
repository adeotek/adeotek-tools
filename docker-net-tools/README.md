# Docker Network Tools Image

Custom Docker image based on Debian 12 (stable-slim) with comprehensive network diagnostic and troubleshooting tools.

## Building the Image

### Using the provided build script

Build without push:

```bash
./build.sh v1.0.0
```

To build and push to the registry:

```bash
./build.sh v1.0.0 --push
```

### Manual build

```bash
docker build -t docker-registry/docker-net-tools:v1.0.0 .
```

## Running the Container

Basic usage:

```bash
docker run -it --rm docker-registry/docker-net-tools:v1.0.0
```

With network host mode (for full network access):

```bash
docker run -it --rm --network host docker-registry/docker-net-tools:v1.0.0
```

## What's Included

### Base Tools
- **micro, nano** - Text editors
- **curl, wget** - HTTP clients
- **mc** - Midnight Commander file manager
- **jq** - JSON processor
- **htop** - Interactive process viewer
- **hstr** - Bash history suggest box with enhanced Ctrl-R
- **git** - Version control
- **bash-completion** - Bash auto-completion

### Network Diagnostic Tools
- **tcpdump** - Packet analyzer
- **tshark** - Network protocol analyzer (Wireshark CLI)
- **nmap** - Network mapper and port scanner
- **ngrep** - Network packet grep
- **iperf, iperf3** - Network bandwidth measurement
- **mtr** - Network diagnostic tool (traceroute + ping)
- **fping** - Fast ping utility
- **traceroute, tcptraceroute** - Network path tracing
- **netcat-traditional** - TCP/IP swiss army knife
- **socat** - Multipurpose relay
- **dnsutils** - DNS tools (dig, nslookup, etc.)
- **apache2-utils** - Apache utilities (ab for benchmarking)
- **httpie** - User-friendly HTTP client
- **speedtest-cli** - Internet speed test

### Network Configuration Tools
- **iproute2** - Advanced IP routing utilities (ip, ss, etc.)
- **bridge-utils** - Ethernet bridge administration
- **ethtool** - Network interface card settings
- **iptables, nftables** - Firewall administration
- **ipset** - IP sets administration
- **conntrack** - Connection tracking tools
- **ipvsadm** - IPVS administration

### System Tools
- **strace, ltrace** - System call and library call tracers
- **openssh-client** - SSH client
- **ca-certificates** - Common CA certificates
- **perl** - Perl interpreter

### Python Tools
- **python3-pip** - Python package installer
- **python3-scapy** - Packet manipulation library

### Libraries
- **libssl3** - SSL/TLS libraries
- **liboping0** - ICMP ping library

## Custom Bash Configuration

The image includes custom bash configuration with:
- **`ll` alias** - Enhanced `ls -lAFh` with colors
- **HSTR (hh)** - Enhanced command history with Ctrl-R binding
- Increased history size (10,000 commands)
- History synchronization between sessions

## Base Image

- **Base:** mirror.gcr.io/library/debian:stable-slim
- **OS:** Debian 12 (Bookworm)
- **Shell:** Bash

## Files in This Directory

- **Dockerfile** - Docker image definition
- **build.sh** - Build script with push support
- **.dockerignore** - Files excluded from build context
- **README.md** - This file

## Notes

- The image runs as root user with working directory `/root`
- Permissions are configured for OpenShift compatibility
- tshark dumpcap is configured with proper permissions
