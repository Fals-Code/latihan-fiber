# syntax=docker/dockerfile:1
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app .

FROM alpine:3.24
RUN addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=build /out/app ./app
RUN mkdir logs && chown app:app logs
USER app
EXPOSE 3000
ENTRYPOINT ["./app"]
CMD ["-role=api"]
