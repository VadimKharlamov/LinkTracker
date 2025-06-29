# Запуск приложения

Для запуска приложения необходимо создать .env файл и указать следующие переменные

BOT_TOKEN = токен от телеграм бота выдается при регистрации бота в BotFather

GITHUB_TOKEN = гитхаб токен для работы с апишкой, получается в инструментах разработчика

STACK_OVERFLOW_TOKEN = ключ от Stack Exchange API

Также по желанию можно изменить адерса по которым работают сервисы. По умолчанию они на localhost и портах 33031 и 33032

# Запуск базы данных и миграций

Для запуска базы данных достаточно выполнить команду docker-compose up -d в этом случае поднимется контейнер с бд и выполнятся liquibase миграции.

Для запуска по отдельности можно использовать docker-compose up postgresql -d | docker-compose up liquibase-migrations -d | docker-compose up redis -d | docker-compose up zookeeper -d | docker-compose up kafka -d
