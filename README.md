# ThinGet

A minimal NuGet package cache proxy for local development. Sits in your docker-compose stack and eliminates upstream latency for previously downloaded packages.

## Quick Start

Add to your `docker-compose.yml`:

```yaml
services:
  nuget-cache:
    image: ghcr.io/shcizo/thinget:latest
    ports:
      - "5555:5555"
    volumes:
      - nuget-cache:/cache

volumes:
  nuget-cache:
```

## Usage with .NET Docker Builds

### 1. Create `nuget.docker.config` in your project

```xml
<configuration>
  <packageSources>
    <clear />
    <add key="thinget" value="http://nuget-cache:5555/v3/index.json" />
  </packageSources>
</configuration>
```

### 2. Add a build arg to your Dockerfile

```dockerfile
FROM mcr.microsoft.com/dotnet/sdk:9.0 AS build
ARG NUGET_CONFIG=nuget.config
WORKDIR /src
COPY ${NUGET_CONFIG} /nuget.config
COPY *.csproj .
RUN dotnet restore --configfile /nuget.config
COPY . .
RUN dotnet publish -c Release -o /app
```

### 3. Wire it up in docker-compose

```yaml
services:
  nuget-cache:
    image: ghcr.io/shcizo/thinget:latest
    volumes:
      - nuget-cache:/cache

  my-app:
    build:
      context: .
      args:
        NUGET_CONFIG: nuget.docker.config
    depends_on:
      - nuget-cache

volumes:
  nuget-cache:
```

The first `docker compose build` fetches packages from nuget.org through the cache. Subsequent builds serve cached packages from disk instantly.

In CI/CD pipelines, don't pass the build arg — the default `nuget.config` is used and ThinGet is bypassed entirely.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `THINGET_PORT` | `5555` | Listen port |
| `THINGET_UPSTREAM` | `https://api.nuget.org` | Upstream NuGet source |
| `THINGET_CACHE_DIR` | `/cache` | Cache directory |

## How It Works

- `GET /v3/index.json` — Service index pointing package downloads to ThinGet, metadata to upstream
- `GET /v3/flat/{id}/index.json` — Version listing, always proxied to upstream
- `GET /v3/flat/{id}/{version}/{file}.nupkg` — Served from cache on hit, fetched and cached on miss
- `GET /health` — Health check

Packages are immutable in NuGet, so cached files never need invalidation. The cache persists in a Docker volume. Clear it with `docker volume rm` if needed.
