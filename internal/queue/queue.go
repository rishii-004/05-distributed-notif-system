package queue

import "github.com/i-katta/notif-system/internal/model"

var EventQueue = make(chan model.Event, 100)
