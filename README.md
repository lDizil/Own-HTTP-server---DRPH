# DRPH — Dizi request processing HTTP (HTTP server)

HTTP/1.1 сервер написанный с нуля на Go поверх сырых TCP соединений без использования `net/http`.

Сделан для изучения внутреннего устройства HTTP и использования как библиотеку для написания веб сервера.

## Возможности

- Роутер с path параметрами (`/users/:id`) и query параметрами
- Все HTTP методы: GET, POST, PUT, PATCH, DELETE
- Middleware — глобальные и на группах маршрутов
- Группировка маршрутов с наследованием middleware
- Gzip сжатие ответов
- Persistent connections (HTTP keep-alive)
- Graceful shutdown — ждёт завершения активных запросов
- Таймауты: 30s idle, 10s на запрос
- Защита от кривых запросов (400) и слишком большого тела (413)
- Настраиваемый `MaxBodySize` (по умолчанию 10 МБ)

## Установка

```
go get github.com/lDizil/Own-HTTP-server---DRPH
```

## Использование

```go
package main

import (
    h "github.com/lDizil/Own-HTTP-server---DRPH"
)

func main() {
    s := h.NewServer()

    s.Use(h.Logger)
    s.Use(h.Recovery)

    s.Get("/", func(ctx *h.Context) {
        ctx.Text(200, "hello")
    })

    s.Get("/users/:id", func(ctx *h.Context) {
        ctx.JSON(200, map[string]string{"id": ctx.Params["id"]})
    })

    s.Post("/users", func(ctx *h.Context) {
        ctx.JSON(201, map[string]string{"status": "created"})
    })

    s.Run("8080")
}
```

## Группировка маршрутов

```go
api := s.Group("/api")
api.Use(AuthMiddleware)

v1 := api.Group("/v1")
v1.Get("/users", listUsers)
v1.Post("/users", createUser)
```

## Свой middleware

```go
func AuthMiddleware(ctx *h.Context, next h.HandlerFunc) {
    token := ctx.Headers["Authorization"]
    if token == "" {
        ctx.Text(401, "unauthorized")
        return
    }
    next(ctx)
}
```

## Настройки

```go
s := h.NewServer()
s.MaxBodySize = 1 << 20  // 1 МБ вместо 10 МБ по умолчанию
```
