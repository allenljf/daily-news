#!/bin/sh
set -eu

exec python -m app.jobs.daily_news "$@"
