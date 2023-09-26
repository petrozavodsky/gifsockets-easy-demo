package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"image"
	"image/color"
	"image/gif"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Ошибка загрузки .env файла:", err)
		return
	}
	http.HandleFunc(os.Getenv("ROUTE"), handleRequest)
	err = http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		return
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {

	timeoutStr := os.Getenv("TIMEOUT_SECONDS")

	seconds, err := time.ParseDuration(timeoutStr + "s")

	if err != nil {
		log.Fatal("Ошибка при преобразовании строки в тип time.Duration:", err)
		return
	}
	duration := time.Duration(seconds.Seconds()) * time.Second
	fmt.Println(duration)

	ctx := r.Context()
	ctx, cancel := context.WithTimeout(r.Context(), duration)
	defer cancel()

	done := make(chan bool)
	start := time.Now()

	// Функция, которая будет выполняться в отдельной goroutine
	go func() {
		fmt.Println("Запрос стартовал")
		gifHandler(w)
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
		webPing(duration.String(), r)
	}
}

func webPing(duration string, r *http.Request) {
	url := os.Getenv("WEB_PING_URL")
	client := http.Client{Timeout: 5 * time.Second}

	//TODO тут нужно формировать url
	resp, err := client.Get(url)
	if err != nil {
		fmt.Println("Ошибка при выполнении запроса:", err)
		return
	}
	defer func(Body io.ReadCloser) {

		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Вебпинг произошел")
		}
	}(resp.Body)
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

func gifHandler(w http.ResponseWriter) {
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
		_, err = w.Write(buffer.Bytes())
		if err != nil {
			return
		}

		time.Sleep(delay * time.Millisecond)
	}
}

func createFrame() *image.Paletted {

	rect := image.Rect(0, 0, 1, 1)
	img := image.NewPaletted(rect, color.Palette{color.White, color.Black})

	img.SetColorIndex(1, 1, 1)

	return img
}
