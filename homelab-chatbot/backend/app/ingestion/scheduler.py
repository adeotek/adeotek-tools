"""Background scheduler that runs the orchestrator at a fixed interval."""

import logging
from pathlib import Path
from typing import Callable

from apscheduler.schedulers.asyncio import AsyncIOScheduler

from app.config import AppConfig
from app.ingestion.orchestrator import IngestionOrchestrator

logger = logging.getLogger(__name__)


class SyncScheduler:
    """Schedules periodic runs of the IngestionOrchestrator."""

    def __init__(
        self,
        config: AppConfig,
        clone_root: Path,
        get_token: Callable[[str], str | None],
    ) -> None:
        self.orchestrator = IngestionOrchestrator(config, clone_root, get_token)
        self._interval = config.sync.interval_seconds
        self._scheduler = AsyncIOScheduler()

    def start(self) -> None:
        self._scheduler.add_job(
            self._safe_run,
            trigger="interval",
            seconds=self._interval,
            id="ingest",
            max_instances=1,
            coalesce=True,
            next_run_time=None,
        )
        self._scheduler.start()
        try:
            self.orchestrator.run_once()
        except Exception as e:  # noqa: BLE001
            logger.exception("initial ingest failed: %s", e)

    def _safe_run(self) -> None:
        try:
            self.orchestrator.run_once()
        except Exception as e:  # noqa: BLE001
            logger.exception("scheduled ingest failed: %s", e)

    def shutdown(self) -> None:
        self._scheduler.shutdown(wait=False)
