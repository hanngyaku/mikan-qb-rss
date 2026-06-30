FROM node:24-alpine AS web-build
WORKDIR /src/web
COPY web/package*.json ./
RUN npm ci
COPY web .
RUN npm run build

FROM golang:1.25-alpine AS go-build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.22
RUN adduser -D -u 1000 app
WORKDIR /app
COPY --from=go-build /server /app/server
COPY --from=web-build /src/web/dist /app/web
RUN mkdir /app/data && chown -R app /app
USER app
EXPOSE 8081
ENV WEB_DIR=/app/web
CMD ["/app/server"]
