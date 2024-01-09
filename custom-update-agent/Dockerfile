FROM --platform=$BUILDPLATFORM golang:alpine AS builder
RUN apk update && apk add --no-cache git
RUN apk update && apk add --no-cache ca-certificates
WORKDIR /build

COPY . .
RUN go mod download

ARG TARGETOS TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o app .

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
WORKDIR /source
COPY . .
WORKDIR /bin
COPY --from=builder /build/app .
WORKDIR /source
CMD ["/bin/app"]