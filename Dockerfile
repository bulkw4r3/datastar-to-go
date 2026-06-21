FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
RUN CGO_ENABLED=0 go build -o numberstore main.go
# Create an empty data directory so the final image can own it as a non-root user.
RUN mkdir /app/data

FROM scratch
COPY --from=builder /app/numberstore /numberstore
COPY --from=builder --chown=1000:1000 /app/data /data
USER 1000:1000
WORKDIR /data
EXPOSE 7085
ENV PORT=7085
ENTRYPOINT ["/numberstore"]
