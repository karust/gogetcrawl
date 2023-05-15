# Build
FROM golang:alpine as builder

LABEL stage=gobuilder
RUN apk update --no-cache && apk add --no-cache tzdata

WORKDIR /build

ADD go.mod .
ADD go.sum .
RUN go mod download

COPY . .
RUN go build -o /app/gogetcrawl .

#ENTRYPOINT ["/app/gogetcrawl/build/gogetcrawl"]

# Run
FROM alpine:3.17

RUN apk update --no-cache && apk add --no-cache ca-certificates
COPY --from=builder /app/gogetcrawl /usr/local/bin/gogetcrawl

RUN adduser \
    --gecos "" \
    --disabled-password \
    gogetcrawl

USER gogetcrawl
ENTRYPOINT ["gogetcrawl"]