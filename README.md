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

Start the cache before building:

```bash
docker compose up nuget-cache -d
```

## Usage with .NET Docker Builds

### 1. Create `nuget.docker.config` in your project

```xml
<configuration>
  <packageSources>
    <clear />
    <add key="thinget" value="http://host.docker.internal:5555/v3/index.json" allowInsecureConnections="true" />
  </packageSources>
  <packageSourceMapping>
    <packageSource key="thinget">
      <package pattern="*" />
    </packageSource>
  </packageSourceMapping>
</configuration>
```

You should also have a default `nuget.config` for non-Docker builds (CI, local `dotnet restore`, etc.):

```xml
<configuration>
  <packageSources>
    <clear />
    <add key="nuget.org" value="https://api.nuget.org/v3/index.json" />
  </packageSources>
  <packageSourceMapping>
    <packageSource key="nuget.org">
      <package pattern="*" />
    </packageSource>
  </packageSourceMapping>
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
    ports:
      - "5555:5555"
    volumes:
      - nuget-cache:/cache

  my-app:
    build:
      context: .
      args:
        NUGET_CONFIG: nuget.docker.config

volumes:
  nuget-cache:
```

The first `docker compose build` fetches packages from nuget.org through the cache. Subsequent builds serve cached packages from disk instantly.

In CI/CD pipelines, don't pass the build arg — the default `nuget.config` is used and ThinGet is bypassed entirely.

## Usage Outside Docker

If you want to use ThinGet for local `dotnet restore` without Docker builds, run the cache container and point your NuGet client at it.

### 1. Start ThinGet

```bash
docker compose up -d nuget-cache
```

### 2. Add a local NuGet source

```bash
dotnet nuget add source http://localhost:5555/v3/index.json -n thinget
```

Or create/edit a `nuget.config` in your solution root:

```xml
<configuration>
  <packageSources>
    <clear />
    <add key="thinget" value="http://localhost:5555/v3/index.json" allowInsecureConnections="true" />
  </packageSources>
</configuration>
```

Then run `dotnet restore` as usual. Packages are fetched through ThinGet and cached for subsequent restores.

To revert, remove the source:

```bash
dotnet nuget remove source thinget
```

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

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| `NuGet requires HTTPS sources` | NuGet blocks HTTP by default | Add `allowInsecureConnections="true"` on the source in `nuget.docker.config` |
| `NU1507: There are 2 package sources` | Central Package Management requires source mapping when multiple sources exist | Add `<packageSourceMapping>` with `<package pattern="*" />` to your NuGet configs |
| `Unable to load the service index` | ThinGet isn't running | Run `docker compose up nuget-cache -d` before building |
| Build can't resolve `nuget-cache` hostname | Docker build containers don't have access to the compose network | Use `host.docker.internal:5555` instead of `nuget-cache:5555` in `nuget.docker.config` |
