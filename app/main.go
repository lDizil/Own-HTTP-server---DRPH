// Package main является точкой входа в программу.
// Также содержит три дополнительных файл помимо main.go
//
// Context.go реализует сценарий формирования полного ответа на конкретный запрос к серверу и отправляет сообщение клиенту.
// На данный момент содержатся следующие основный методы: bytes - главный метод, который подготавливает ответ, остальные являются обёрткой для него (вызывают и передают нужные параметры)
// Text, JSON, File - обёркт над bytes, конкретизируют ответ в зависимости от запроса и роутера
//
// Router.go реализует слой работы с роутерами: добавление существующих путей и дальнейшее сравнение пути запроса с добавленными.
// Также формирует path параметры (map[string]string), которые в дальнейшем используются для создания контекста.
// Формирование path именно в этом слое, а не в context,
// так как path параметры - побочный продукт процесса поиска маршрута. Match их получает бесплатно в процессе своей основной работы.
//
// Server.go реализует слой открытия соединения и парсинга всех запросов. Создаёт context запроса, не закрывает соединение до EOF или заголовка
// Connection = close.
package main

import (
	"fmt"
	"os"

	h "github.com/lDizil/Own-HTTP-server---DRPH/httpserver"
)

func main() {
	fmt.Println("Запуск http сервера")

	args := os.Args
	var dirName string
	for i, arg := range args {
		if arg == "--directory" {
			dirName = args[i+1]
		}
	}

	s := h.NewServer()

	s.Use(h.Recovery)
	s.Use(h.Logger)

	api := s.Group("/api") 
	{
		test := api.Group("/test")

		test.Get("/", func(ctx *h.Context) {
			ctx.Text(200, "")
		})
	}
	s.Get("/", func(ctx *h.Context) {
		ctx.Text(200, "")
	})

	s.Get("/echo/:str", func(ctx *h.Context) {
		ctx.Text(200, ctx.Params["str"])
	})

	s.Get("/user-agent", func(ctx *h.Context) {
		ctx.Text(200, ctx.Headers["User-Agent"])
	})

	s.Get("/files/:filename", func(ctx *h.Context) {
		fileName := ctx.Params["filename"]

		data, err := os.ReadFile(dirName + "/" + fileName)

		if err != nil {
			ctx.Text(404, "")
			return
		}

		ctx.File(200, data)
	})

	s.Post("/files/:filename", func(ctx *h.Context) {
		fileName := ctx.Params["filename"]

		err := os.WriteFile(dirName+"/"+fileName, ctx.Body, 0644)

		if err != nil {
			ctx.Text(500, "")
			return
		}

		ctx.File(201, []byte{})
	})

	s.Run("4221")
}
