FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.22
RUN adduser -D app
WORKDIR /app
COPY --from=build /server /app/server
RUN mkdir /app/data && chown -R app /app
USER app
EXPOSE 8081
CMD ["/app/server"]
