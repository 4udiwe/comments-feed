# Comments Feed - GraphQL System

Система для добавления и чтения постов и комментариев с использованием GraphQL.

## Возможности

- Создание и просмотр постов
- Иерархические комментарии с неограниченной вложенностью
- Запрет на комментирование для автора поста
- Пагинация постов и комментариев (в том числе вложенных)
- Поддержка in-memory и PostgreSQL хранилищ

## Запуск
Склонировать репозиторий
```bash
git clone https://github.com/4udiwe/comments-feed
cd comments-feed
```

### Быстрый запуск в in-memory режиме (go run)

```bash
# Запустить с in-memory хранилищем
STORAGE_TYPE=memory go run cmd/main.go
```

### С использованием Docker Compose

```bash
# Запуск в режиме хранения in-memory
make up-memory

# Запуск с хранением в Postgres
make up-postgres

# Остановка и логи
make down
make logs
```

### Переменные окружения
Переменная `CONFIG_PATH` строго указана в [docker-compose.yaml](docker-compose.yml)

Переменная `STORAGE_TYPE` (memory | postgres) указывается в [Makefile](Makefile) при запуске приложения или вручную при запуске через go run

Остальные параметры, такие как уровень логирования, можно указать в [config/config.yaml](config/config.yaml).


## Использование

### GraphQL Playground

По умолчанию доступен на `http://localhost:8080/`

<details>
<summary>Примеры запросов</summary>

### Создание поста

```graphql
mutation CreatePost {
  createPost(input: {
    title: "simple post"
    content: "simple post with enabled comments"
    commentsEnabled: true
  }) {
    id
    title
    content
    commentsEnabled
    createdAt
  }
}
```

### Создание root комментария

```graphql
mutation CreateComment {
  createComment(input: {
    postId: "8f460b53-a231-4855-a113-82206e459f47"
    text: "simple comment to simple post"
  }) {
    id
    text
    parentId
  }
}
```

### Создание ответа на комментарий

```graphql
mutation ReplyComment {
  createComment(input: {
    postId: "8f460b53-a231-4855-a113-82206e459f47"
    parentId: "3d994cba-f8a4-48ab-b7e4-f91dd9a412c4"
    text: "simple reply to simple comment to simple post"
  }) {
    id
    text
    parentId
  }
}
```

### Получение постов с комментариями и ответами (с пагинацией)
```graphql
query GetPosts {
  posts(limit: 10, offset: 0) {
    id
    title
    content
    comments(limit: 10, offset: 0) {
      id
      text
      children(limit: 10, offset: 0) {
        id
        text
      }
    }
  }
}
```

### Получение поста с комментариями и ответами (с пагинацией)

```graphql
query GetPost {
  post(id: "8f460b53-a231-4855-a113-82206e459f47") {
    id
    title
    comments(limit: 10, offset: 0) {
      id
      text
      createdAt
      children(limit: 10, offset: 0){
        id
        text
      }
    }
  }
}
```
</details>

### Подписка на комментарии
Реализована подписка на новые комментарии конкретного поста.

<details>
<summary>Пример подписки</summary>

```graphql
subscription SubscribeToComments{
  commentAdded(postId: "77699861-c395-4dec-bc1f-5a6ea8a02a6d") {
    text
  }
}
```
</details>


## Тестирование

### Unit тесты

В проекте реализовано unit-тестирование сервисного слоя. Для запуска тестов можно воспользоваться командой:
```bash
make test
```

## Оптимизация

### Индексы в БД

 - `idx_comment_post_id` - Для получения комментариев к посту
 - `idx_comment_parent_id` - Для получения ответов на комментарий
 - `idx_comment_post_id_parent_id` - Для получения всех комментариев с ответами

В качестве аналога индекса и ускорения поиска для in-memory хранилища были использованы map, хранящие комментарии и посты

```golang
type CommentRepository struct {
	mu sync.RWMutex

	byID       map[string]*model.Comment
	byPostID   map[string][]*model.Comment
	byParentID map[string][]*model.Comment
}
```

### N+1 
Для решения проблемы множества подзапросов были использованы:
- [DataLoader](internal/graph/loaders/loader.go) - загружает комментарии батчем для всех постов сразу
- Batch-методы для получения данных в репозиториях

### Логирование и обработка ошибок
Для логирования входящих HTTP-запросов используется [http_logging_middleware](internal/middleware/http_logging.go).

Для логирования GrahpQL-запросов используется [operation_logging_middleware](internal/graph/middleware/logging.go) и [resolver_logging_middleware](internal\graph\middleware\resolver_logging.go) для логирования ошибок в resolver'ах.

Для обработки паник при запросе используется [resolver_logging_middleware](internal/graph/middleware/resolver_logging.go).