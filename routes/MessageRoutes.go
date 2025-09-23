package routes

import (
	"github.com/gorilla/mux"

	"HeatingEventServiceGo/controllers"
)

func SetupRoutes(router *mux.Router) {
	router.HandleFunc("/messages", controllers.CreateMessage).Methods("POST")
	router.HandleFunc("/messages", controllers.GetAllMessages).Methods("GET")
	router.HandleFunc("/messages/filter", controllers.FilterMessages).Methods("GET")
}
