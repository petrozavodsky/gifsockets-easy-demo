package main

import (
	"bytes"
	"context"
	"github.com/joho/godotenv"
	"image"
	"image/color"
	"image/gif"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Panic("Ошибка загрузки .env файла:", err)
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

	ctx := r.Context()
	ctx, cancel := context.WithTimeout(r.Context(), duration)
	defer cancel()

	done := make(chan bool)
	start := time.Now()

	// Функция, которая будет выполняться в отдельной goroutine
	go func() {
		log.Println("Запрос стартовал")
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
		webPing(duration, r)
	}
}

func webPing(duration time.Duration, r *http.Request) {

	timeStr := strconv.Itoa(int(duration.Seconds()))
	url := os.Getenv("WEB_PING_URL")
	url = strings.Replace(url, "{TIME}", timeStr, -1)
	client := http.Client{Timeout: 5 * time.Second}
	args := parseArgs(r)

	if len(args) > 0 {
		for key, value := range args {
			url = strings.Replace(url, key, value, -1)
		}
	}
	resp, err := client.Get(url)
	if err != nil {
		log.Fatal("Ошибка при выполнении запроса:", err)
		return
	}
	defer func(Body io.ReadCloser) {

		err := Body.Close()
		if err != nil {
			log.Fatal(err)
		} else {
			log.Println("Выполнен вебпинг - ", url)
		}
	}(resp.Body)
}

func parseArgs(r *http.Request) map[string]string {
	queryParams := r.URL.Query()
	args := make(map[string]string)

	if len(queryParams) < 1 {
		return args
	}
	for key, values := range queryParams {
		args[key] = strings.Join(values, " ")
	}
	return args
}

func gifHandler(w http.ResponseWriter) {
	const delay = 1000 // Задержка между кадрами в миллисекундах

	for {
		img := createFrame()

		buffer := new(bytes.Buffer)
		err := gif.Encode(buffer, img, nil)

		if err != nil {
			log.Fatal(err)
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
