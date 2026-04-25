# Networking

## VLAN Layout

- VLAN 10: Management (switches, routers, IPMI)
- VLAN 20: Services (app servers, containers)
- VLAN 30: IoT (sensors, cameras)

## Firewall

OPNsense runs on a mini PC. Rules block inter-VLAN traffic by default.
