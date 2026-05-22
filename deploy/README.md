# Развёртывание стенда (Makefile)

Из **корня репозитория** VKR:

| Цель | Команда |
|------|---------|
| Поднять **тестовый** полный стек (pytest / ручная проверка, проект `vkr-test`) | `make test-up` |
| Остановить тестовый стек | `make test-down` |
| Поднять **прод-подобный** стек (тот же compose, проект `vkr-prod`) | `make prod-up` |
| Остановить прод-подобный стек | `make prod-down` |
| Логи | `make test-logs` / `make prod-logs` |

Оба стека используют один файл [`sed-service/e2e_tests/docker-compose.test.yml`](../sed-service/e2e_tests/docker-compose.test.yml); отличается только имя Docker-проекта (`vkr-test` / `vkr-prod`). **Хостовые порты одинаковые (28xxx)** — на одной машине не запускайте `test-up` и `prod-up` одновременно.

Прод (`make prod-up`): один внешний порт **gateway** — `API_GATEWAY_PORT` в корневом `.env` (по умолчанию **52556** → внутри контейнера **5656**). Публичные URL: `API_PUBLIC_URL`, `KEYCLOAK_PUBLIC_URL` с тем же хостом и портом. На роутере проброс **52556→192.168.1.199:52556** (не 5656).
