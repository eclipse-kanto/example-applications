FROM --platform=$BUILDPLATFORM golang:alpine AS builder
RUN apk update && apk add --no-cache git
WORKDIR /build
COPY . .
RUN go mod download

ARG TARGETOS TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o app .
RUN mkdir ./fileagent

FROM scratch
WORKDIR /tmp
COPY --from=builder /tmp .
WORKDIR /bin
COPY --from=builder /build/fileagent .
COPY --from=builder /build/app .
CMD ["/bin/app"]