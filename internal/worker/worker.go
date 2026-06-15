package worker

import (
	"log"

	"github.com/i-katta/notif-system/internal/queue"
)

func Start(count int) {
	for i := range count {
		go func(id int) {
			for event := range queue.EventQueue {
				log.Printf("Worker %d processed event: %s", id, event.EventType)
			}
		}(i)
	}
}
