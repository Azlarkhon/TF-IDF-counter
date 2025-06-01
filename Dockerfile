FROM golang:1.24.3 AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Статическая сборка бинарника
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o /bin/main ./main.go

FROM alpine:3.18 AS production

# Копируем бинарник
COPY --from=build /bin/main /bin/main

# Копируем всё остальное из проекта, кроме кода (нужные файлы)
COPY --from=build /src /app
WORKDIR /app

EXPOSE 8080
CMD ["/bin/main"]
