"""Regenerate the integration-test Excel fixture. Run manually when the schema changes."""

from pathlib import Path

import pandas as pd

HERE = Path(__file__).parent


def main() -> None:
    devices = pd.DataFrame(
        [
            {"hostname": "nas-01", "role": "storage", "cpu_cores": 8, "ram_gb": 32},
            {"hostname": "proxmox-01", "role": "hypervisor", "cpu_cores": 16, "ram_gb": 64},
            {"hostname": "pihole-01", "role": "dns", "cpu_cores": 2, "ram_gb": 4},
        ]
    )
    with pd.ExcelWriter(HERE / "sample.xlsx", engine="openpyxl") as w:
        devices.to_excel(w, sheet_name="devices", index=False)


if __name__ == "__main__":
    main()
