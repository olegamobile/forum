FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o forum

# Using the debian:bookworm-slim image for a small runtime environment
FROM debian:bookworm-slim
WORKDIR /app

# Metadata
LABEL project="Forum"
LABEL description="A web forum made with golang and SQLite3 at localhost:8080"

COPY --from=builder /app/forum /app/forum
COPY --from=builder /app/templates /app/templates
COPY --from=builder /app/static /app/static
COPY --from=builder /app/forum.db /app/forum.db

# Expose port 8080 to allow the app to be accessible from outside the container
EXPOSE 8080
CMD ["/app/forum"]
