package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"gorm.io/gorm/clause"

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

func CreateMessage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Cообщение");
    var req CreateMessageRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        fmt.Println("Ошибка декодирования тела запроса:", err)
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    message := models.Message{
        Message:    req.Message,
        Severity:   req.Severity,
        Source:     req.Source,
        ServerTime: req.ServerTime,
        SourceTime: req.SourceTime,
    }
    if message.ServerTime.IsZero() {
        message.ServerTime = time.Now()
    }
    if message.SourceTime.IsZero() {
        message.SourceTime = time.Now()
    }

    if err := models.DB.Create(&message).Error; err != nil {
        fmt.Println("Ошибка создания сообщения в БД:", err)
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Правильный вызов для рассылки всем клиентам
    // BroadcastTo отправляет сообщение всем клиентам в указанном namespace и room.
    // Пустая строка для "room" означает рассылку всем клиентам в этом namespace.
    if socket.Server != nil {
        fmt.Println("Отправка события new-message через Socket.IO:", message)
        socket.Server.BroadcastToRoom("/", "", "new-message", message)
    } else {
        fmt.Println("Socket.IO сервер не инициализирован, событие не отправлено")
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"message": "Message Created"})
}

func GetAllMessages(w http.ResponseWriter, r *http.Request) {
	var messages []models.Message
	if err := models.DB.Find(&messages).Error; err != nil {
		fmt.Println("Ошибка получения всех сообщений:", err)
		http.Error(w, `{"message": "INTERNAL SERVER ERROR"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Println("Возвращено сообщений:", len(messages))
	json.NewEncoder(w).Encode(messages)
}

func FilterMessages(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	limitStr := query.Get("limit")
	var limit *int
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err == nil {
			limit = &l
		}
	}

	offsetStr := query.Get("offset")
	var offset *int
	if offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err == nil {
			offset = &o
		}
	}

	source := query.Get("source")
	beginStr := query.Get("begin")
	endStr := query.Get("end")

	var begin, end *time.Time
	if beginStr != "" {
		b, err := time.Parse(time.RFC3339, beginStr)
		if err == nil {
			begin = &b
		}
	}
	if endStr != "" {
		e, err := time.Parse(time.RFC3339, endStr)
		if err == nil {
			end = &e
		}
	}

	dbQuery := models.DB.Model(&models.Message{})

	if source != "" {
		dbQuery = dbQuery.Where("source = ?", source)
	}

	if begin != nil && end != nil {
		dbQuery = dbQuery.Where("serverTime BETWEEN ? AND ?", begin, end)
	} else if begin != nil {
		dbQuery = dbQuery.Where("serverTime >= ?", begin)
	} else if end != nil {
		dbQuery = dbQuery.Where("serverTime <= ?", end)
	}

	dbQuery = dbQuery.Order(clause.OrderByColumn{Column: clause.Column{Name: "serverTime"}, Desc: true})

	if limit != nil {
		dbQuery = dbQuery.Limit(*limit)
	}
	if offset != nil {
		dbQuery = dbQuery.Offset(*offset)
	}

	var messages []models.Message
	if err := dbQuery.Find(&messages).Error; err != nil {
		fmt.Println("Ошибка запроса к БД в FilterMessages:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Println("FilterMessages возвращено сообщений:", len(messages))
	json.NewEncoder(w).Encode(messages)
}
