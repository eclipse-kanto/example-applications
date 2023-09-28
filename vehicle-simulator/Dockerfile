FROM --platform=$BUILDPLATFORM golang:alpine AS builder
RUN apk update && apk add --no-cache git

WORKDIR /build

COPY . .
RUN go mod download

ARG TARGETOS TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o app .

FROM scratch
WORKDIR /source
COPY . .
WORKDIR /bin
COPY --from=builder /build/app .
CMD ["/bin/app"]
