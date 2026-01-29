FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache make git protobuf-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o bin/sink ./cmd/sink
RUN CGO_ENABLED=0 go build -o bin/sensor ./cmd/sensor

FROM builder AS tester
RUN go install gotest.tools/gotestsum@latest
CMD ["gotestsum", "--format", "testname", "--", "-v", "./..."]

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/sink .
COPY --from=builder /app/bin/sensor .

RUN mkdir -p /app/logs

EXPOSE 50051

CMD ["./sink"]
