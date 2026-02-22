package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/ws"
)

type WSHandler struct {
	hub            *ws.Hub
	upgrader       websocket.Upgrader
	allowedOrigins map[string]struct{}
}

func NewWSHandler(hub *ws.Hub, allowedOrigins []string) *WSHandler {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[origin] = struct{}{}
	}

	return &WSHandler{
		hub:            hub,
		allowedOrigins: allowed,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return false
				}
				_, ok := allowed[origin]
				return ok
			},
		},
	}
}

func (h *WSHandler) ServeWS(c *gin.Context) {
	userID := c.GetString(string(models.ContextKeyUserID))
	familyID := c.GetString(string(models.ContextKeyFamilyID))

	if userID == "" || familyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication is required or your session is invalid."})
		return
	}

	origin := c.GetHeader("Origin")
	if origin == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Request origin is not allowed."})
		return
	}
	if _, ok := h.allowedOrigins[origin]; !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Request origin is not allowed."})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		respondInternalErrorWithMessage(c, "Unable to establish websocket connection at the moment.", err)
		return
	}

	client := ws.NewClient(h.hub, conn, userID, familyID)
	go client.Run()
}
