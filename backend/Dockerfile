FROM golang:1.27-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY migrations ./migrations
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/api ./cmd/api \
    && CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/daily-news-job ./cmd/daily-news-job

FROM alpine:3.21

WORKDIR /app
RUN addgroup -S daily-news && adduser -S -G daily-news -h /app daily-news
COPY --from=build /out/api /app/api
COPY --from=build /out/daily-news-job /app/daily-news-job
COPY migrations /app/migrations
USER daily-news

EXPOSE 8080

CMD ["/app/api"]
