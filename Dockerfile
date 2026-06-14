ARG GO_IMAGE=docker.m.daocloud.io/library/golang:1.26-alpine
ARG MIHOMO_IMAGE=docker.io/metacubex/mihomo:v1.19.25
ARG RUNTIME_IMAGE=docker.m.daocloud.io/library/alpine:latest
ARG METACUBEXD_REPO=https://github.com/MetaCubeX/metacubexd.git
ARG METACUBEXD_REF=be93782fc673ece6689eb135437196d7359e27a9

FROM docker.m.daocloud.io/library/node:22-bookworm-slim AS metacubexd_fork_builder

ARG METACUBEXD_REPO
ARG METACUBEXD_REF
ARG METACUBEXD_GIT_PROXY

WORKDIR /metacubexd
RUN sed -i \
      -e 's#http://deb.debian.org/debian#http://mirrors.aliyun.com/debian#g' \
      -e 's#http://deb.debian.org/debian-security#http://mirrors.aliyun.com/debian-security#g' \
      /etc/apt/sources.list.d/debian.sources \
    && apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates git \
    && rm -rf /var/lib/apt/lists/*
RUN git init \
    && git remote add origin "${METACUBEXD_REPO}" \
    && if [ -n "${METACUBEXD_GIT_PROXY}" ]; then \
         git \
           -c http.proxy="${METACUBEXD_GIT_PROXY}" \
           -c https.proxy="${METACUBEXD_GIT_PROXY}" \
           -c http.lowSpeedLimit=1000 \
           -c http.lowSpeedTime=60 \
           fetch --depth 1 origin "${METACUBEXD_REF}"; \
       else \
         git \
           -c http.lowSpeedLimit=1000 \
           -c http.lowSpeedTime=60 \
           fetch --depth 1 origin "${METACUBEXD_REF}"; \
       fi \
    && git checkout --detach FETCH_HEAD
RUN npm config set registry https://repo.huaweicloud.com/repository/npm/ \
    && npm install -g pnpm@10.34.1 \
    && pnpm config set registry https://repo.huaweicloud.com/repository/npm/ \
    && pnpm config set fetch-timeout 600000 \
    && HUSKY=0 pnpm install --frozen-lockfile \
    && pnpm add country-region-data@4.1.0 --save-exact
COPY metacubexd-fork ./metacubexd-fork
RUN for patch in metacubexd-fork/patches/*.patch; do git apply "${patch}"; done \
    && for dir in pages components composables types; do \
         if [ -d "metacubexd-fork/${dir}" ]; then cp -R "metacubexd-fork/${dir}/." "${dir}/"; fi; \
       done \
    && NUXT_APP_BASE_URL=/ pnpm generate

FROM ${GO_IMAGE} AS builder

WORKDIR /app

ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY gen ./gen
COPY internal ./internal
RUN go build -o proxy-runtime ./cmd/proxy-runtime

FROM ${MIHOMO_IMAGE} AS mihomo

FROM ${RUNTIME_IMAGE} AS mihomo_extract
COPY --from=mihomo / /mihomo-root
RUN set -eux; \
    for candidate in /mihomo-root/mihomo /mihomo-root/usr/bin/mihomo /mihomo-root/usr/local/bin/mihomo /mihomo-root/bin/mihomo; do \
      if [ -f "$candidate" ]; then cp "$candidate" /mihomo; chmod +x /mihomo; exit 0; fi; \
    done; \
    found="$(find /mihomo-root -type f -name mihomo | head -n 1)"; \
    test -n "$found"; \
    cp "$found" /mihomo; chmod +x /mihomo

FROM ${RUNTIME_IMAGE}

WORKDIR /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=mihomo_extract /mihomo /usr/local/bin/mihomo
COPY --from=builder /app/proxy-runtime /usr/local/bin/proxy-runtime
COPY --from=metacubexd_fork_builder /metacubexd/.output/public /app/dashboard/metacubexd

EXPOSE 8080 1080

CMD ["proxy-runtime"]
