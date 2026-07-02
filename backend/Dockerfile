FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod ./
COPY main.go ./
RUN go build -o splitbill-server .

FROM alpine:3.20
WORKDIR /app
COPY --from=build /app/splitbill-server ./splitbill-server
EXPOSE 8080
ENTRYPOINT ["./splitbill-server"]
