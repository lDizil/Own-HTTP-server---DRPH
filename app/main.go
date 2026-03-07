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

	r := &h.Router{}
	s := h.NewServer(r)

	s.Use(h.Recovery)
	s.Use(h.Logger)

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

		err := os.WriteFile(dirName + "/"+ fileName, ctx.Body, 0644)

		if err != nil {
			ctx.Text(500, "")
			return
		}

		ctx.File(201, []byte{})
	})

	fmt.Println("Http сервер успешно запущен")

	s.Listen("0.0.0.0:4221")

}

