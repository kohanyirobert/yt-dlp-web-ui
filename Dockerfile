# Node (pnpm) ------------------------------------------------------------------
FROM node:22-slim AS ui
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack prepare pnpm@10.0.0 --activate && corepack enable
COPY . /usr/src/yt-dlp-web-ui

WORKDIR /usr/src/yt-dlp-web-ui/frontend

RUN rm -rf node_modules

RUN pnpm install
RUN pnpm run build
# -----------------------------------------------------------------------------

# Go --------------------------------------------------------------------------
FROM golang AS build

WORKDIR /usr/src/yt-dlp-web-ui

COPY . .
COPY --from=ui /usr/src/yt-dlp-web-ui/frontend /usr/src/yt-dlp-web-ui/frontend

RUN CGO_ENABLED=0 GOOS=linux go build -o yt-dlp-web-ui
# -----------------------------------------------------------------------------

# Runtime ---------------------------------------------------------------------
FROM python:3.13.2-alpine3.21

RUN apk update && \
apk add ffmpeg ca-certificates curl wget gnutls --no-cache && \
pip install "yt-dlp[default,curl-cffi,mutagen,pycryptodomex,phantomjs,secretstorage]"

VOLUME /downloads /config

WORKDIR /app

COPY --from=build /usr/src/yt-dlp-web-ui/yt-dlp-web-ui /app

ENV JWT_SECRET=secret

EXPOSE 3033
ENTRYPOINT [ "./yt-dlp-web-ui" , "--out", "/downloads", "--conf", "/config/config.yml", "--db", "/config/local.db" ]
