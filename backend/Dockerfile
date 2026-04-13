FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN export GOPROXY=https://goproxy.cn,direct && export GOSUMDB=sum.golang.google.cn && go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o deeplx-web ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates python3 py3-pip

WORKDIR /app

COPY --from=builder /app/requirements.txt .
RUN pip3 install --no-cache-dir --break-system-packages -r requirements.txt

COPY --from=builder /app/deeplx-web .
COPY --from=builder /app/scripts ./scripts

RUN mkdir -p uploads logs

EXPOSE 8449

CMD ["./deeplx-web"]
