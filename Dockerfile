FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
RUN CGO_ENABLED=0 go build -o numberstore main.go

FROM scratch
COPY --from=builder /app/numberstore /numberstore
EXPOSE 8080
ENV PORT=8080
ENTRYPOINT ["/numberstore"]
