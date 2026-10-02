FROM golang:1.22 as builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-extldflags "-static"' -o pciedatabase .

# Certificate generation stage
# NOTE: self-signed cert, valid 365 days from build time. It is only used for the
# ingress-nginx -> pod hop, and the ingress is configured with proxy-ssl-verify "false"
# (see deploy/gpu-database.yaml), so expiry does not cause an outage.
FROM debian:bookworm-slim as cert-generator
RUN apt-get update && \
    apt-get install -y openssl && \
    rm -rf /var/lib/apt/lists/* && \
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout /key.pem -out /cert.pem \
    -subj "/C=US/ST=State/L=City/O=PCIeDatabase/CN=pciedatabase"

# Final runtime image
FROM gcr.io/distroless/base-debian12

# Copy binary and certificates
COPY --from=builder /app/pciedatabase /
COPY --from=cert-generator /cert.pem /
COPY --from=cert-generator /key.pem /

EXPOSE 443

CMD ["/pciedatabase"]
