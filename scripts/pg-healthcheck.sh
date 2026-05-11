#!/bin/sh
# Healthcheck para Postgres usando variables de entorno
set -eu
pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB"
