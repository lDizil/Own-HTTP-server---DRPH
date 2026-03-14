// Пример использования http сервера DRPH
package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"time"

	h "github.com/lDizil/Own-HTTP-server---DRPH"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
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
	s.Use(h.Metrics)

	api := s.Group("/api")
	{
		test := api.Group("/test")

		test.Use(func(ctx *h.Context, next h.HandlerFunc) {})

		test.Get("/", func(ctx *h.Context) {
			ctx.Text(200, "")
		})
	}

	s.Get("/metrics", func(ctx *h.Context) {
		mfs, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			ctx.Text(500, err.Error())
			return
		}

		var buf bytes.Buffer
		enc := expfmt.NewEncoder(&buf, expfmt.NewFormat(expfmt.TypeTextPlain))
		for _, mf := range mfs {
			enc.Encode(mf)
		}

		ctx.SetResponseHeader("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		ctx.Bytes(200, buf.Bytes())
	})

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

	s.Get("/slow", func(ctx *h.Context) {
		ms := rand.Intn(450) + 50
		time.Sleep(time.Duration(ms) * time.Millisecond)
		ctx.Text(200, fmt.Sprintf("done in %dms", ms))
	})

	s.Run("8080")
}
