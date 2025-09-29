package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"HeatingEventServiceGo/config"
	"HeatingEventServiceGo/models"
	"HeatingEventServiceGo/routes"
	"HeatingEventServiceGo/socket"

	"github.com/rs/cors"

	"github.com/gorilla/mux"
	socketio "github.com/zishang520/socket.io/servers/socket/v3"
)

func main() {
	// Подключение к БД
	config.ConnectDB()
	models.DB = config.DB

	// Создаём Socket.IO сервер
	socket.Server = socketio.NewServer(nil, nil)

	// Обработка подключения клиентов
	socket.Server.On("connection", func(clients ...any) {
		sock := clients[0].(*socketio.Socket)
		log.Printf("✅ Клиент подключён, ID: %s", sock.Id())

		// Отправка события "connected" клиенту
		sock.Emit("connected", map[string]interface{}{
			"message": "Успешное подключение к серверу",
			"id":      sock.Id(),
			"time":    time.Now().Format("15:04:05"),
		})

		// Обработка события "message"
		sock.On("message", func(data ...any) {
			if len(data) > 0 {
				log.Printf("📨 Сообщение от %s: %v", sock.Id(), data[0])
				sock.Emit("message_reply", fmt.Sprintf("Сообщение доставлено: %v", data[0]))
				// Рассылаем всем клиентам
				socket.Server.Emit("new-message", map[string]interface{}{
					"from":    sock.Id(),
					"message": data[0],
					"time":    time.Now().Format("15:04:05"),
				})
			}
		})

		// Обработка тестового события "test"
		sock.On("test", func(data ...any) {
			if len(data) > 0 {
				log.Printf("🧪 Тестовое сообщение от %s: %v", sock.Id(), data[0])
				sock.Emit("test_reply", fmt.Sprintf("Тест пройден: %v", data[0]))
			}
		})

		// Отключение клиента
		sock.On("disconnect", func(reason ...any) {
			r := "unknown"
			if len(reason) > 0 {
				r = fmt.Sprintf("%v", reason[0])
			}
			log.Printf("👋 Клиент отключён, ID: %s, Причина: %s", sock.Id(), r)
		})
	})

	// HTTP маршруты
	router := mux.NewRouter()
	routes.SetupRoutes(router)

	// Эндпоинт для проверки здоровья
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"SocketIOService"}`))
	})

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},                             // Разрешает все origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "OPTIONS"}, // Добавили OPTIONS явно
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Content-Length"},
		AllowCredentials: true,
		Debug:            true, // Включаем дебаг rs/cors (логи в консоль сервера)
	})

	handler := c.Handler(router)
	// Привязываем Socket.IO к HTTP
	http.Handle("/socket.io/", socket.Server.ServeHandler(nil))
	http.Handle("/", c.Handler(handler))

	// Запуск сервера
	port := os.Getenv("PORT")
	if port == "" {
		port = "5200"
	}
	log.Printf("Сервер запущен на порту %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
