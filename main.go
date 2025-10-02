package main

import (
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

	socket.Server = socketio.NewServer(nil, opts)

	socket.Server.On("connection", func(clients ...any) {
		sock := clients[0].(*socketio.Socket)
		log.Printf("Клиент подключён, ID: %s", sock.Id())
	})

	router := mux.NewRouter()
	routes.SetupRoutes(router)

	router.PathPrefix("/socket.io/").Handler(socket.Server.ServeHandler(nil))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},                             // Разрешает все origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "OPTIONS"}, // Добавили OPTIONS явно
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		Debug:            false, // Включаем дебаг rs/cors (логи в консоль сервера)
	}).Handler(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5200"
	}
	log.Printf("Сервер запущен на порту %s", port)
	log.Fatal(http.ListenAndServe(":"+port, c))
}
