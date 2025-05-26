FROM golang:1.24.3 AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Статическая сборка бинарника
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/hello ./main.go

FROM alpine AS production

# Копируем бинарник
COPY --from=build /bin/hello /bin/main

# Копируем всё остальное из проекта, кроме кода
COPY --from=build /src /app
WORKDIR /app

EXPOSE 8080
CMD ["/bin/main"]
