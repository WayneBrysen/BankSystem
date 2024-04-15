# Build stage
FROM golang:1.21-alpine3.19 AS builder
WORKDIR /app
COPY . .

# set go module download
RUN go env -w GOPROXY=https://goproxy.io,direct

RUN go build -o main main.go

# Run stage
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/main .
COPY app.env .
COPY start.sh .
COPY wait-for .
COPY db/migration ./db/migration

EXPOSE 8080
CMD [ "/app/main" ]
ENTRYPOINT [ "/app/start.sh" ]