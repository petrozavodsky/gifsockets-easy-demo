package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	http.HandleFunc("/pixel.gif", handleRequest)
	http.ListenAndServe(":82", nil)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	done := make(chan bool)
	start := time.Now()

	// Функция, которая будет выполняться в отдельной goroutine
	go func() {
		fmt.Println("Запрос стартовал")
		gifHandler(w, r)
		done <- true
	}()

	select {
	case <-done:
		// Ответ клиенту
		w.WriteHeader(http.StatusOK)
	case <-ctx.Done():
		// Обработка разрыва соединения
		duration := time.Since(start)
		showResult(duration.String(), r)
	}
}

func showResult(duration string, r *http.Request) {
	queryParams := r.URL.Query()

	message := time.Now().Format("2006-01-02 15:04:05")
	message += " время визита "
	message += duration
	if len(queryParams) > 0 {
		message += " GET-параметры запроса: "
	}
	for key, values := range queryParams {
		message += strings.Join([]string{key, "=>", strings.Join(values, " "), " "}, " ")
	}

	fmt.Println(message)
}

func gifHandler(w http.ResponseWriter, r *http.Request) {
	const delay = 1000 // Задержка между кадрами в миллисекундах

	for {
		img := createFrame()

		buffer := new(bytes.Buffer)
		err := gif.Encode(buffer, img, nil)

		if err != nil {
			log.Println(err)
			return
		}

		w.Header().Set("Content-Type", "image/gif")
		w.Write(buffer.Bytes())

		time.Sleep(delay * time.Millisecond)
	}
}

func createFrame() *image.Paletted {

	rect := image.Rect(0, 0, 1, 1)
	img := image.NewPaletted(rect, color.Palette{color.White, color.Black})

	img.SetColorIndex(1, 1, 1)

	return img
}
