package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"HeatingEventServiceGo/models"
	"HeatingEventServiceGo/socket"
)

type CreateMessageRequest struct {
	Message    string    `json:"message"`
	Severity   int       `json:"severity"`
	Source     string    `json:"source"`
	ServerTime time.Time `json:"serverTime,omitempty"`
	SourceTime time.Time `json:"sourceTime,omitempty"`
}

// POST /messages
func CreateMessage(w http.ResponseWriter, r *http.Request) {
	var req CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	msg := models.Message{
		Message:    req.Message,
		Severity:   req.Severity,
		Source:     req.Source,
		ServerTime: req.ServerTime,
		SourceTime: req.SourceTime,
	}

	if msg.ServerTime.IsZero() {
		msg.ServerTime = time.Now()
	}
	if msg.SourceTime.IsZero() {
		msg.SourceTime = time.Now()
	}

	if err := models.DB.Create(&msg).Error; err != nil {
		fmt.Println("Ошибка создания сообщения:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем событие всем подключённым клиентам через Socket.IO
	if socket.Server != nil {
		socket.Server.Emit("new-message", msg)
		fmt.Println("Событие new-message отправлено всем клиентам")
	} else {
		fmt.Println("Socket.IO сервер не инициализирован")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Message Created"})
}

// GET /messages
func GetAllMessages(w http.ResponseWriter, r *http.Request) {
	var messages []models.Message
	if err := models.DB.Find(&messages).Error; err != nil {
		http.Error(w, `{"message":"INTERNAL SERVER ERROR"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

// GET /messages/filter?source=&begin=&end=&limit=&offset=
func FilterMessages(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	source := query.Get("source")
	beginStr := query.Get("begin")
	endStr := query.Get("end")
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

	dbQuery := models.DB.Model(&models.Message{})

	if source != "" {
		dbQuery = dbQuery.Where("source = ?", source)
	}

	if beginStr != "" {
		if begin, err := time.Parse(time.RFC3339, beginStr); err == nil {
			dbQuery = dbQuery.Where("server_time >= ?", begin)
		}
	}
	if endStr != "" {
		if end, err := time.Parse(time.RFC3339, endStr); err == nil {
			dbQuery = dbQuery.Where("server_time <= ?", end)
		}
	}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			dbQuery = dbQuery.Limit(l)
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			dbQuery = dbQuery.Offset(o)
		}
	}

	var messages []models.Message
	if err := dbQuery.Order("server_time DESC").Find(&messages).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
