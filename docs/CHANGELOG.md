# Changelog

Все важные изменения в этом проекте будут документироваться в этом файле.

## Feedback

* От [@Jack23DTX](https://github.com/Jack23DTX):

```
Инструкция по запуску понятна, запускал контейнеры на локальной машине, с проблемами не столкнулся.

на VM не смог запустить, ибо нет доступа к ним :(

Понравилось:
+ использование cors для фронтенда;
+ обратный прокси на nginx;
+ использование JWT для аутентификации;

Скорее не хватило нежели, чем не понравилось:
- мало комментариев в коде;
- возможность загружать несколько файлов сразу;
- минимальные тесты по подсчетам tf, idf, huffman, для проверки их "правильности";
- параллельная обработка подсчетов (например, одновременно вычислять tf и idf);


Скорее как плюс, для кодировки Хаффмана использовать кэширование (меньше потребления памяти и меньше нагрузка на cpu).
```

## Types of Changes

* `Added` — для новых функций
* `Changed` — для изменений существующей функциональности
* `Deprecated` — для функций, которые скоро будут удалены
* `Removed` — для функций, которые уже удалены
* `Fixed` — для исправлений ошибок
* `Security` — в случае уязвимостей
* `Dependency` — для обновлений зависимостей
* `Performance` — для улучшений производительности
* `Experimental` — для экспериментальных функций

### [11.06.2025] — v1.2.0

### Added

* Авторизация и Аутентификация (cookie based JWT)
* API для пользователей (login, register, logout, updatePassword, delete user)
* API для документов и коллекций (crud и указанные в тз)
* Volume для хранения документов (user_id/document)
* GET /documents/:document_id/statistics — дает стату так, словно все документы из всех коллекций(где находится document_id)) находятся в одной коллекции.
* Swagger ui
* Диаграмма датабазы
* Интерфейсы для контроллеров
* CORS middleware
* Закодирвание данных документа с помощью алгоритма Хаффмана

### Changed

* Добавил транзакцию(db) в HandleFileUpload
* Архитектура проекта

### Dependency

* Добавлен github.com/golang-jwt/jwt/v4 для генерации jwt
* Добавлен golang.org/x/crypto для хэширования пароля юзера
* Добавлен github.com/swaggo/gin-swagger
* Добавлен github.com/swaggo/files

### Removed

* templates - фронтенд мне уже не нужен, т.к. имеется сваггер документация

## [29.05.2025] — v1.1.0

### Added

* `.env` — конфигурация переменных окружения
* `config/init.go` — инициализация конфигурации
* `helper/responseBuilder.go` — универсальный формат ответов API
* `version/version.go` — хранение версии приложения
* `controllers/systemParametersController.go` — эндпоинты `/status` и `/version`
* `Dockerfile` и `compose.yaml` — контейнеризация приложения
* База данных PostgreSQL и модели (`models/`)
* Эндпоинт `/metrics` для мониторинга и сбора статистики:
  * `files_processed`
  * `latest_file_processed_timestamp`
  * `min_time_processed`, `avg_time_processed`, `max_time_processed`
  * `total_file_size_mb`, `avg_file_size_mb`
  * `top_10_most_freq_words`
* `README.md` — описание проекта
* Nginx — базовая настройка для проксирования, ограничения нагрузки и масштабируемости

### Dependency

* `github.com/joho/godotenv v1.5.1` — поддержка `.env`
* `gorm.io/driver/postgres` — PostgreSQL-драйвер для GORM
* `gorm.io/gorm` — ORM
* Обновление версии Go: `v1.24.0` → `v1.24.3`
