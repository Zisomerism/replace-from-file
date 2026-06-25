FROM golang:1.26 AS builder

WORKDIR /app
COPY . /app

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -trimpath -v -o app .

FROM gcr.io/distroless/static

COPY --from=builder /app/app /app

ENTRYPOINT ["/app"]
