# Greenlight

## Description

Greenlight is a simple movie API playground that I use to try out new 
concepts with Go.

## Running

1. Run `docker compose up -d` to start core stack (Postgres and Redis)
2. Set `GREENLIGHT_DB_DSN` env variable
3. Perform DB migrations with `make db/migrations/up`
4. Run the API with `make run/api`
