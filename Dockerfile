FROM --platform=$BUILDPLATFORM golang:latest as build
ARG TARGETOS TARGETARCH
ENV CGO_ENABLED=0
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN --mount=target=. \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /out/routeros-webhook .

FROM scratch

COPY <<EOF /etc/passwd
root:x:0:0:root:/root:/bin/sh
unprivileged:x:1000:1000::/home/unprivileged:/bin/sh
EOF
COPY --from=build /out/routeros-webhook /routeros-webhook
USER unprivileged

ENTRYPOINT [ "/routeros-webhook" ]
