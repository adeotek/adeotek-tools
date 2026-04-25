from app.ingestion.embed import Embedder


def test_embedder_returns_384_dims():
    e = Embedder(model_name="BAAI/bge-small-en-v1.5")
    vecs = e.embed_batch(["hello world", "another sentence"])
    assert len(vecs) == 2
    assert len(vecs[0]) == 384
    assert len(vecs[1]) == 384


def test_embedder_deterministic():
    e = Embedder(model_name="BAAI/bge-small-en-v1.5")
    v1 = e.embed_batch(["home lab"])[0]
    v2 = e.embed_batch(["home lab"])[0]
    assert v1 == v2
