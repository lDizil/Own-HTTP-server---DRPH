package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	fmt.Println("Запуск http сервера")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	// Чтение аргументов командой строки и парсинг флагов
	args := os.Args

	var dirName string

	for ind, arg := range args {
		if arg == "--directory" {
			dirName = args[ind+1]
		}
	}

	for {
		conn, err := l.Accept()

		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConn(conn, dirName)

	}
}

func handleConn(conn net.Conn, dirName string) {

	reader := bufio.NewReader(conn)

	// Чтение http запроса, который приходит на сервер. Его парсинг на начальную строку, заголовки
	for {
		headers := make(map[string]string)


		pathLine, err := reader.ReadString('\n')

		if err != nil {
			conn.Close()
			break
		}

		path := strings.Split(pathLine, " ")

		for {

			headersLine, _ := reader.ReadString('\n')

			if headersLine == "\r\n" {
				break
			}

			reqPartsHeaders := strings.Split(headersLine, ": ")
			headers[reqPartsHeaders[0]] = strings.TrimSpace(reqPartsHeaders[1])

		}
		

		// Обработка флага закрытия (для закрытия conn и корректного ответа)
		var shouldClose bool

		if val, ok := headers["Connection"]; ok {
			if val == "close" {
				shouldClose = true
			}
		}

		var body []byte
		if lenStr, ok := headers["Content-Length"]; ok {
			length, _ := strconv.Atoi(strings.TrimSpace(lenStr))
			body = make([]byte, length)
			io.ReadFull(reader, body)
		}

		answer, bodyAnswer := route(path, headers, body, dirName)

		answer = addConnectionClose(answer, shouldClose)

		conn.Write([]byte(answer))

		if bodyAnswer != nil {
			conn.Write(bodyAnswer)
		}

		if shouldClose {	
			conn.Close()
			break
		}
	}
}

// Составление ответа на запрос, учитывая распарсенные данные
func route(path []string, headers map[string]string, body []byte, dirName string) (string, []byte) {

	answer := ""
	var bodyAnswer []byte

	if path[1] == "/" {
		answer = "HTTP/1.1 200 OK\r\n\r\n"
		return answer, bodyAnswer

	} else if echoStr, ok := strings.CutPrefix(path[1], "/echo/"); ok {

		if encodes, ok := headers["Accept-Encoding"]; ok {

			diffEncodes := strings.Split(encodes, ", ")

			for _, encode := range diffEncodes {
				if encode == "gzip" {

					var buf bytes.Buffer

					gzipWriter := gzip.NewWriter(&buf)
					_, err := gzipWriter.Write([]byte(echoStr))

					if err != nil {
						answer = "HTTP/1.1 500 Internal Server Error\r\n\r\n"
						return answer, bodyAnswer
					}

					gzipWriter.Close()

					bodyAnswer = buf.Bytes()

					answer = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Encoding: gzip\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n", len(bodyAnswer))
					return answer, bodyAnswer
				}
			}

			if answer == "" {
				answer = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(echoStr), echoStr)
				return answer, bodyAnswer
			}

		} else {
			answer := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(echoStr), echoStr)
			return answer, bodyAnswer
		}

	} else if path[1] == "/user-agent" {
		val := headers["User-Agent"]
		answer = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(val), val)
		return answer, bodyAnswer

	} else if pathToFile, ok := strings.CutPrefix(path[1], "/files/"); ok {
		switch path[0] {

		case "GET":
			file, err := os.ReadFile(dirName + "/" + pathToFile)

			if err != nil {
				answer = "HTTP/1.1 404 Not Found\r\n\r\n"
				return answer, bodyAnswer
			} else {
				answer = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(file), string(file))
				return answer, bodyAnswer
			}

		case "POST":
			err := os.WriteFile(dirName+"/"+pathToFile, body, 0644)

			if err != nil {
				answer = "HTTP/1.1 500 Internal Server Error\r\n\r\n"
				return answer, bodyAnswer
			} else {
				answer = "HTTP/1.1 201 Created\r\n\r\n"
				return answer, bodyAnswer
			}

		}

	}

	answer = "HTTP/1.1 404 Not Found\r\n\r\n"
	return answer, bodyAnswer

}


func addConnectionClose(response string, shouldClose bool) string {
	if shouldClose {
		response = strings.Replace(response, "\r\n\r\n", "\r\nConnection: close\r\n\r\n", 1)
	}
	
	return response
}