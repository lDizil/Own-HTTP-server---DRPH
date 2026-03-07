package main

import (
	"fmt"
	"os"
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

	r := &Router{}
	s := NewServer(r, dirName)

	s.Get("/", func(ctx *Context) {
		ctx.Text(200, "")
	})
	
	s.Get("/echo/:str", func(ctx *Context) {
		ctx.Text(200, ctx.Params["str"])
	})

	s.Get("/user-agent", func(ctx *Context) {
		ctx.Text(200, ctx.Headers["User-Agent"])
	})

	s.Get("/files/:filename", func(ctx *Context) {
		fileName := ctx.Params["filename"]

		data, err := os.ReadFile(s.dirName + "/" + fileName)

		if err != nil {
			ctx.Text(404, "")
			return
		}

		ctx.File(200, data)
	})

	s.Post("/files/:filename", func(ctx *Context) {
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

