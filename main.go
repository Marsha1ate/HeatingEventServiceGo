// main.go

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	
	socketio "github.com/googollee/go-socket.io"

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

	server := socketio.NewServer(nil)

	socket.Server = server

	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		fmt.Println("Клиент подключён, ID:", s.ID())
		return nil
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		fmt.Println("Ошибка сокета:", e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		fmt.Println("Клиент отключён, причина:", reason)
	})

	go func() {
		if err := server.Serve(); err != nil {
			log.Fatalf("Ошибка запуска Socket.IO: %s\n", err)
		}
	}()
	defer server.Close()

	router := mux.NewRouter()
	routes.SetupRoutes(router)

	mainMux := http.NewServeMux()
	mainMux.Handle("/socket.io/", server)
	mainMux.Handle("/", router)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		Debug: true,
	})
	handler := c.Handler(mainMux)

	fmt.Printf("Сервер запущен на порту %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}