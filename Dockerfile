FROM golang:1.22 as builder

WORKDIR /app

COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o wrapper-manager

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/wrapper-manager .
RUN chmod +x ./wrapper-manager

ENTRYPOINT ["./wrapper-manager"]
EXPOSE 8080