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
	"github.com/zishang520/socket.io/v3/pkg/types"
)

func main() {
	// Подключение к БД
	config.ConnectDB()
	models.DB = config.DB

	opts := socketio.DefaultServerOptions()
	opts.SetPingTimeout(20 * time.Second)
	opts.SetPingInterval(25 * time.Second)
	opts.SetMaxHttpBufferSize(1e6)
	opts.SetCors(&types.Cors{
		Origin:      "*",
		Credentials: true,
	})

	// Создаём сервер Socket.IO
	socket.Server = socketio.NewServer(nil, opts)

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

	router.PathPrefix("/socket.io/").Handler(socket.Server.ServeHandler(nil))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},                             // Разрешает все origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "OPTIONS"}, // Добавили OPTIONS явно
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		Debug:            false, // Включаем дебаг rs/cors (логи в консоль сервера)
	}).Handler(router)

	// Привязываем Socket.IO к HTTP
	/*http.Handle("/socket.io/", c)
	http.Handle("/", c)
	muxRouter := http.NewServeMux()
	muxRouter.Handle("/socket.io/", socket.Server.ServeHandler(nil))*/
	/*muxHandler := http.NewServeMux()
	muxHandler.Handle("/socket.io/", c)
	muxHandler.Handle("/", cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}).Handler(router))*/

	//http.ListenAndServe(":5200", muxRouter)

	// Запуск сервера
	port := os.Getenv("PORT")
	if port == "" {
		port = "5200"
	}
	log.Printf("Сервер запущен на порту %s", port)
	log.Fatal(http.ListenAndServe(":"+port, c))
}
