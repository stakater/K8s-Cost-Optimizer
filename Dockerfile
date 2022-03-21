FROM golang:1.17.8 as builder

ARG TARGETOS
ARG TARGETARCH
ARG GOPROXY
ARG GOPRIVATE

WORKDIR /build

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY pkg/ pkg/
COPY main.go ./

RUN CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    GOPROXY=${GOPROXY} \
    GOPRIVATE=${GOPRIVATE} \
    GO111MODULE=on \
    go build -mod=mod -a -o k8s-cost-optimizer main.go

# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

FROM scratch

USER 65532:65532

COPY --from=builder /build/k8s-cost-optimizer /bin/k8s-cost-optimizer

ENTRYPOINT ["/bin/k8s-cost-optimizer"]
