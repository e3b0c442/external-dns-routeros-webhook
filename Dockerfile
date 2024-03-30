FROM --platform=$BUILDPLATFORM golang:latest as build
ARG TARGETOS TARGETARCH
ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN --mount=target=. \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /out/routeros-webhook .

FROM scratch

COPY --from=build /out/routeros-webhook /routeros-webhook

ENTRYPOINT [ "/routeros-webhook" ]
