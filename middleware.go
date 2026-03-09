package httpserver

import (
	"fmt"
	"time"
)

func Logger(ctx *Context, next HandlerFunc) {
	start := time.Now()

	method := ctx.Method

	path := ctx.Path

	next(ctx)

	elapsed := time.Since(start).Microseconds()

	nowTime := time.Now().Format("02/Jan/2006:15:04:05 -0700")
	
	//формат логов: 192.168.1.1 - - [10/Mar/2023:12:00:01 +0000] "GET /index.html HTTP/1.1" 200 (время выполнения: 20)
	fmt.Printf("%s - - [%s] \"%s %s HTTP/1.1\" %d (время выполнения: %v мк)\n", ctx.RemoteAddr, nowTime, method, path, ctx.StatusCode(), elapsed)
}

func Recovery(ctx *Context, next HandlerFunc) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("panic:", r)
            ctx.Text(500, "")
		}
	}()

	next(ctx)
}