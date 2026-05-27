# SSH Tunnel

> 13 nodes

## Key Concepts

- **SSHTunnel** (11 connections) — `internal/tunnel/tunnel.go`
- **.establish()** (5 connections) — `internal/tunnel/tunnel.go`
- **.getAuthMethod()** (4 connections) — `internal/tunnel/tunnel.go`
- **NewSSHTunnel()** (3 connections) — `internal/tunnel/tunnel.go`
- **.forward()** (3 connections) — `internal/tunnel/tunnel.go`
- **.handleConnection()** (3 connections) — `internal/tunnel/tunnel.go`
- **.Close()** (3 connections) — `internal/tunnel/tunnel.go`
- **tunnel.go** (2 connections) — `internal/tunnel/tunnel.go`
- **.getKeyAuth()** (2 connections) — `internal/tunnel/tunnel.go`
- **.getAgentAuth()** (2 connections) — `internal/tunnel/tunnel.go`
- **.GetLocalAddress()** (1 connections) — `internal/tunnel/tunnel.go`
- **.GetLocalPort()** (1 connections) — `internal/tunnel/tunnel.go`
- **.HealthCheck()** (1 connections) — `internal/tunnel/tunnel.go`

## Relationships

- [[CLI App Commands]] (1 shared connections)

## Source Files

- `internal/tunnel/tunnel.go`

## Audit Trail

- EXTRACTED: 40 (98%)
- INFERRED: 1 (2%)
- AMBIGUOUS: 0 (0%)

---

*Part of the graphify knowledge wiki. See [[index]] to navigate.*