package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/i-katta/notif-system/internal/model"
	"github.com/i-katta/notif-system/internal/queue"
)

func PostEvent(c *gin.Context) {
	var event model.Event
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	select {
	case queue.EventQueue <- event:
	default:
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "queue full"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"status": "accepted", "event_id": event.EventID})
}
