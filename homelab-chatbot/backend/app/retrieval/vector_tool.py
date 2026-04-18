"""Vector-search tool exposed to the LLM agent."""

from app.ingestion.embed import Embedder
from app.storage.lance import SearchHit, VectorStore


class VectorSearchTool:
    """Embed a query and return top-K matching markdown chunks."""

    TOOL_NAME = "search_homelab_docs"

    def __init__(
        self,
        store: VectorStore,
        embedder: Embedder,
        top_k: int = 5,
    ) -> None:
        self._store = store
        self._embedder = embedder
        self._default_top_k = top_k

    def search(
        self,
        query: str,
        top_k: int | None = None,
        repo: str | None = None,
    ) -> list[SearchHit]:
        vec = self._embedder.embed_batch([query])[0]
        return self._store.search(
            query_vector=vec, top_k=top_k or self._default_top_k, repo=repo
        )

    def as_llama_tool(self):
        """Return a LlamaIndex FunctionTool wrapping this tool."""
        from llama_index.core.tools import FunctionTool

        def _run(query: str, repo: str | None = None) -> list[dict]:
            """Search the home lab documentation for information."""
            return [
                {
                    "text": hit.text,
                    "repo": hit.repo,
                    "file_path": hit.file_path,
                    "heading_path": hit.heading_path,
                    "line_start": hit.line_start,
                    "line_end": hit.line_end,
                }
                for hit in self.search(query, repo=repo)
            ]

        return FunctionTool.from_defaults(
            fn=_run,
            name=self.TOOL_NAME,
            description=(
                "Search the home lab documentation by semantic similarity. "
                "Use for prose/conceptual questions. Optional `repo` filter narrows to a specific repo."
            ),
        )
