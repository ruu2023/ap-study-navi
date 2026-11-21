# Cross-compile on the build host (no emulation), run on distroless.
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY *.go ./
COPY templates ./templates
COPY static ./static
ARG TARGETOS TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /out/ap-study . \
 && mkdir -p /out/data

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/ap-study /ap-study
# Owned by nonroot so a fresh named volume inherits writable ownership.
COPY --from=build --chown=65532:65532 /out/data /data
USER nonroot:nonroot
ENV DB_PATH=/data/ap-study.db ADDR=:8080 BASE_PATH=/ap TZ=Asia/Tokyo
EXPOSE 8080
ENTRYPOINT ["/ap-study"]
