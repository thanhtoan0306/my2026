FROM golang:1.23-alpine AS build

WORKDIR /src
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/pcmonitor .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=build /out/pcmonitor /app/pcmonitor

ENV PORT=8095
ENV HOST=0.0.0.0
EXPOSE 8095

CMD ["/app/pcmonitor"]

