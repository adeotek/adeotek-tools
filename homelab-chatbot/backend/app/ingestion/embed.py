"""Wrapper around sentence-transformers for deterministic embeddings."""

from functools import lru_cache

from sentence_transformers import SentenceTransformer

EMBED_DIM = 384


class Embedder:
    """Thin wrapper around SentenceTransformer with batch-embedding convenience."""

    def __init__(self, model_name: str, cache_dir: str | None = None) -> None:
        self._model = _load_model(model_name, cache_dir)

    def embed_batch(self, texts: list[str]) -> list[list[float]]:
        result = self._model.encode(
            texts,
            batch_size=32,
            convert_to_numpy=True,
            normalize_embeddings=True,
            show_progress_bar=False,
        )
        vecs = [vec.tolist() for vec in result]
        if vecs and len(vecs[0]) != EMBED_DIM:
            raise ValueError(f"model returned {len(vecs[0])}-dim vectors, expected {EMBED_DIM}")
        return vecs


@lru_cache(maxsize=4)
def _load_model(model_name: str, cache_dir: str | None) -> SentenceTransformer:
    return SentenceTransformer(model_name, cache_folder=cache_dir)
