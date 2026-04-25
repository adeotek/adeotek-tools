"""Background scheduler that runs the orchestrator at a fixed interval."""

import logging

from apscheduler.schedulers.asyncio import AsyncIOScheduler

from app.config import RepoConfig
from app.ingestion.orchestrator import IngestionOrchestrator

logger = logging.getLogger(__name__)


class SyncScheduler:
    """Schedules periodic runs of the IngestionOrchestrator."""

    def __init__(
        self,
        orchestrator: IngestionOrchestrator,
        interval_seconds: int,
    ) -> None:
        self._orchestrator = orchestrator
        self._interval = interval_seconds
        self._scheduler = AsyncIOScheduler()
        self._repos: list[RepoConfig] = []

    def start(self, repos: list[RepoConfig]) -> None:
        self._repos = repos
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
            self._orchestrator.run_once(repos)
        except Exception as e:  # noqa: BLE001
            logger.exception("initial ingest failed: %s", e)

    def _safe_run(self) -> None:
        try:
            self._orchestrator.run_once(self._repos)
        except Exception as e:  # noqa: BLE001
            logger.exception("scheduled ingest failed: %s", e)

    def stop(self) -> None:
        self._scheduler.shutdown(wait=False)
