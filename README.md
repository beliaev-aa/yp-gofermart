[![Coverage Status](https://coveralls.io/repos/github/beliaev-aa/yp-gofermart/badge.svg?branch=master)](https://coveralls.io/github/beliaev-aa/yp-gofermart?branch=master)

# Gophermart Service

Gophermart — это сервис, разработанный на Go, с поддержкой микросервиса Accrual, который взаимодействует с PostgreSQL. Данный проект включает инструкции по сборке и локальному запуску с использованием Docker и Docker Compose.

## Содержание

- [Требования](#требования)
- [Настройка окружения](#настройка-окружения)
- [Сборка и запуск приложения](#сборка-и-запуск-приложения)
- [Тестирование приложения](#тестирование-приложения)
- [Полезные команды Docker](#полезные-команды-Docker)

## Требования

Перед началом работы убедитесь, что на вашем компьютере установлены:

- [Docker](https://www.docker.com)
- [Docker Compose](https://docs.docker.com/compose/)

## Настройка окружения

1. Склонируйте репозиторий проекта на локальную машину:
   ```bash
   git clone <URL вашего репозитория>
   cd <директория проекта>
   ```

2. Создайте файл .env на основе примера env.dist:

   ```bash
   cp env.dist .env
   ```

3. Настройте переменные окружения в файле .env при необходимости. Например, вы можете изменить порты или настройки базы данных.

## Сборка и запуск приложения

Проект использует Docker и Docker Compose для локального запуска. Следуйте этим шагам для сборки и запуска всех необходимых сервисов:

1. Соберите и запустите сервисы с помощью Docker Compose:
   ```bash
   docker-compose up --build
   ```

   Это создаст и запустит следующие сервисы:

   - gophermart: основной сервис на Go.
   - accrual: микросервис Accrual для обработки начислений.
   - db: контейнер с базой данных PostgreSQL.


2. Откройте браузер или используйте инструменты командной строки, такие как curl, чтобы проверить работу приложения:
   Для регистрации пользователя:

   ```bash
   curl -X POST http://localhost:9090/api/user/register -d '{"login": "user", "password": "password"}' -H "Content-Type: application/json"
   ```
   Другие примеры доступных запросов можно посмотреть в [SPECIFICATION.md](./minio/SPECIFICATION.md)

## Тестирование приложения

Чтобы запустить тесты для сервиса gophermart, выполните следующие команды:

1. Остановите текущие контейнеры, если они запущены:
   ```bash
   docker-compose down
   ```
2. Запустите тесты во время сборки контейнера gophermart:
   ```bash
   docker-compose up --build --abort-on-container-exit
   ```
   Тесты будут выполнены в процессе сборки приложения gophermart на этапе RUN go test ./....

## Полезные команды Docker

Вот несколько полезных команд для работы с Docker и Docker Compose:

- Остановить контейнеры и удалить все связанные ресурсы (контейнеры, сети, тома):
   ```bash
   docker-compose down
   ```

- Просмотреть логи всех сервисов:
   ```bash
   docker-compose logs -f
   ```

- Перезапустить конкретный сервис (например, gophermart):
   ```bash
   docker-compose up --build gophermart
   ```
- Проверить статус контейнеров:
   ```bash
   docker ps
   ```
- Подключиться к контейнеру:
   ```bash
   docker exec -it <container_name> /bin/```bash
   ```