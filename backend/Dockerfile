FROM ghcr.io/astral-sh/uv:0.12.5 AS uv

FROM python:3.12-slim

ENV PATH="/app/.venv/bin:${PATH}" \
    PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    UV_COMPILE_BYTECODE=1 \
    UV_LINK_MODE=copy

WORKDIR /app

RUN groupadd --system daily-news \
    && useradd --system --gid daily-news --home-dir /app daily-news

COPY --from=uv /uv /uvx /bin/
COPY pyproject.toml uv.lock ./
RUN uv sync --frozen --no-dev --no-install-project

COPY app ./app
COPY scripts ./scripts
RUN chmod 0555 scripts/serve.sh scripts/run-job.sh \
    && chown --recursive daily-news:daily-news /app

USER daily-news

EXPOSE 8080

CMD ["./scripts/serve.sh"]
