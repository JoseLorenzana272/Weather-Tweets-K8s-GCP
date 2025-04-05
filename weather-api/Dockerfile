FROM rust:alpine AS builder
WORKDIR /app
COPY . .
RUN apk add --no-cache musl-dev && \
    cargo build --release

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/target/release/weather-api .
CMD ["./weather-api"]