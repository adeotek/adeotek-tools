# Storage

## TrueNAS

Primary storage is a TrueNAS Scale box with 6 x 8TB drives in RAIDZ2.
SMB shares are exposed to the LAN; NFS to the services VLAN.

## Backup

Restic backs up to Backblaze B2 nightly.
