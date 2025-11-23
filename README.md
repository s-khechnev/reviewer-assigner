# Reviewer-assigner

## Запуск
```bash
docker-compose up
```
или
```bash
make
```

Сервис будет доступен на порту 8080.

### Тесты
```bash
make test
```


### Инфо

1. Простой эндпоинт статистики - [GET /stats/reviewers/assignments?status=open&active_only](./api/openapi.yml#L446)
2. Конфиг линтера - [.golangci.yaml](.golangci.yaml)
3. [Интеграционные тесты](integration_tests) 
