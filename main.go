package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
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
	defer func() {
		if err := recover(); err != nil {
			log.Println("Возникла ошибка:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}()

	// Прочитать значение переменной типа bool
	verbose, err := strconv.ParseBool(os.Getenv("VERBOSE"))
	if err != nil {
		panic("Ошибка чтения переменной VERBOSE")
	}

	if w == nil {
		panic("Ошибка http.ResponseWriter")
	}

	timeoutStr := os.Getenv("TIMEOUT_SECONDS")
	seconds, err := time.ParseDuration(timeoutStr + "s")

	if err != nil {
		panic("Ошибка при преобразовании строки в тип time.Duration")
	}
	duration := time.Duration(seconds.Seconds()) * time.Second

	ctx := r.Context()
	ctx, cancel := context.WithTimeout(r.Context(), duration)
	defer cancel()

	done := make(chan bool)
	start := time.Now()

	// Функция, которая будет выполняться в отдельной goroutine
	go func() {
		defer func() {
			done <- true
			if err := recover(); err != nil {
				if verbose {
					log.Println("Ошибка выделения памяти", err)
				}
			}
		}()

		if verbose {
			log.Println("Инициализация соединения")
		}
		gifHandler(w, ctx)
	}()

	select {
	case <-done:
		if w == nil {
			log.Println("Ошибка http.ResponseWriter: объект w равен nil")
		}
	case <-ctx.Done():
		// Обработка разрыва соединения
		duration := time.Since(start)
		webPing(duration, r)
	}
}

func webPing(duration time.Duration, r *http.Request) {

	// Прочитать значение переменной типа bool
	verbose, err := strconv.ParseBool(os.Getenv("VERBOSE"))

	if err != nil {
		log.Println("Ошибка чтения переменной VERBOSE")
		return
	}

	seconds := int(duration.Seconds())
	if seconds < 1 {
		seconds = 1
	}

	timeStr := strconv.Itoa(seconds)

	urlStr := os.Getenv("WEB_PING_URL")
	urlStr = strings.Replace(urlStr, "{TIME}", timeStr, -1)
	urlStr = strings.Replace(urlStr, "{userAgent}", url.QueryEscape(r.Header.Get("User-Agent")), -1)
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		urlStr = strings.Replace(urlStr, "{remoteAddr}", url.QueryEscape(ip), -1)
	} else {
		urlStr = strings.Replace(urlStr, "{remoteAddr}", "", -1)
	}

	client := http.Client{Timeout: 5 * time.Second}
	args := parseArgs(r)

	if len(args) > 0 {
		for key, value := range args {
			urlStr = strings.Replace(urlStr, key, url.QueryEscape(value), -1)
		}
	}

	fmt.Println(urlStr)

	resp, err := client.Get(urlStr)
	if err != nil {
		log.Println("Ошибка при выполнении запроса:", err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal(err)
		} else if verbose {
			log.Println("Выполнен вебпинг - ", urlStr)
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

func gifHandler(w http.ResponseWriter, ctx context.Context) {
	timeoutStr := os.Getenv("TIMEOUT_SECONDS")
	seconds, err := strconv.Atoi(timeoutStr)
	if err != nil {
		panic("Ошибка при преобразовании строки в тип time.Duration")
	}

	const delay = 1000 // Задержка между кадрами в миллисекундах
	for {
		select {
		case <-ctx.Done(): // Проверяем состояние контекста
			return
		default:
			img := createFrame()

			buffer := new(bytes.Buffer)
			err := gif.Encode(buffer, img, nil)

			if err != nil {
				log.Fatal(err)
				return
			}

			if w != nil {
				w.Header().Set("Connection", "keep-alive")
				w.Header().Set("Content-Type", "image/gif")
				w.Header().Set("Content-Length", fmt.Sprintf("%d", len(buffer.Bytes())*seconds+1))
				_, err = w.Write(buffer.Bytes())
				if err != nil {
					return
				}
			}

			time.Sleep(delay * time.Millisecond)
		}

	}
}

func createFrame() *image.Paletted {

	rect := image.Rect(0, 0, 1, 1)
	img := image.NewPaletted(rect, color.Palette{color.White, color.Black})

	img.SetColorIndex(1, 1, 1)

	return img
}
