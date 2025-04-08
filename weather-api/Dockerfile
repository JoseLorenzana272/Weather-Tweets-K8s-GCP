FROM rust:alpine AS builder
WORKDIR /app
COPY . .
RUN apk add --no-cache musl-dev protobuf && \
    cargo build --release

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/target/release/weather-api .
EXPOSE 8080
CMD ["./weather-api"]