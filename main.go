// main.go
package main

import (
	"fmt"
	"log"
	"net/http"

	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/polling"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"HeatingEventServiceGo/config"
	"HeatingEventServiceGo/models"
	"HeatingEventServiceGo/routes"
	"HeatingEventServiceGo/socket"
)

func main() {
	config.ConnectDB()
	models.DB = config.DB
	models.DB.AutoMigrate(&models.Message{})

	port := config.GetEnv("PORT", "5200")

	server := socketio.NewServer(&engineio.Options{
		Transports: []transport.Transport{
			&polling.Transport{
				CheckOrigin: func(r *http.Request) bool {
					return true // разрешаем любые origin
				},
			},
			&websocket.Transport{
				CheckOrigin: func(r *http.Request) bool {
					return true // разрешаем любые origin
				},
			},
		},
	})

	server.OnConnect("/", func(s socketio.Conn) error {
		if s == nil {
			fmt.Println("Ошибка: Socket.IO клиент nil при подключении")
			return fmt.Errorf("nil connection")
		}
		socket.CurrentSocket = s
		s.SetContext("")
		fmt.Println("Новый клиент подключён по Socket.IO, ID:", s.ID())
		return nil
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		if s == nil {
			fmt.Println("Ошибка Socket.IO: клиент nil, ошибка:", e)
			return
		}
		fmt.Println("meet error:", e)
		fmt.Println("Ошибка Socket.IO, клиент", s.ID(), ":", e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		if s == nil {
			fmt.Println("Ошибка: Socket.IO клиент nil при отключении, причина:", reason)
			return
		}
		fmt.Println("closed", reason)
		fmt.Println("Клиент Socket.IO отключён, ID:", s.ID(), ", причина:", reason)
		if socket.CurrentSocket != nil && socket.CurrentSocket.ID() == s.ID() {
			socket.CurrentSocket = nil
		}
	})

	go func() {
		if err := server.Serve(); err != nil {
			fmt.Println("Ошибка запуска Socket.IO сервера:", err)
		}
	}()
	defer server.Close()

	router := mux.NewRouter()
	routes.SetupRoutes(router)

	// Добавьте middleware для логирования HTTP-запросов
	router.Use(loggingMiddleware)

	// Расширенная CORS-конфигурация (чтобы точно соответствовать origin: "*")
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},                                                                 // Разрешает все origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "OPTIONS", "DELETE", "PATCH"},                  // Добавили OPTIONS явно
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Origin"}, // Разрешаем нужные заголовки
		ExposedHeaders:   []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           300,  // Кэширование preflight на 5 мин
		Debug:            true, // Включаем дебаг rs/cors (логи в консоль сервера)
	})

	muxMain := http.NewServeMux()
	muxMain.Handle("/socket.io/", server)
	muxMain.Handle("/", c.Handler(router))

	fmt.Printf("Сервер запущен на порту %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// Middleware для логирования запросов
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Получен запрос: %s %s от %s\n", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
