package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const requestIDContextKey = "request_id"

func respondInternalError(c *gin.Context, err error) {
	respondInternalErrorWithMessage(c, internalErrorMessage(c.Request.Method, c.FullPath()), err)
}

func respondInternalErrorWithMessage(c *gin.Context, message string, err error) {
	if err != nil {
		_ = c.Error(err)
	}

	body := gin.H{
		"error": message,
	}
	if reqID := c.GetString(requestIDContextKey); reqID != "" {
		body["request_id"] = reqID
	}
	c.JSON(http.StatusInternalServerError, body)
}

func internalErrorMessage(method, path string) string {
	key := method + " " + path
	if msg, ok := internalErrorMessages[key]; ok {
		return msg
	}
	return "The server could not process your request at this time. Please try again."
}

var internalErrorMessages = map[string]string{
	"POST /v1/auth/signup":                         "Unable to create account at the moment.",
	"PATCH /v1/auth/me/password":                   "Unable to update password at the moment.",
	"GET /v1/family":                               "Unable to load family details at the moment.",
	"PATCH /v1/family":                             "Unable to update family details at the moment.",
	"GET /v1/family/members":                       "Unable to load family members at the moment.",
	"POST /v1/family/members":                      "Unable to create family member at the moment.",
	"PATCH /v1/family/members/:id":                 "Unable to update family member at the moment.",
	"GET /v1/family/members/:id/rant-count":        "Unable to retrieve rant count at the moment.",
	"GET /v1/family/access-control":                "Unable to load access control settings at the moment.",
	"PUT /v1/family/access-control/:grantee_id":    "Unable to update access control settings at the moment.",
	"DELETE /v1/family/access-control/:grantee_id": "Unable to revoke access control settings at the moment.",
	"GET /v1/hadiths":                              "Unable to load hadiths at the moment.",
	"GET /v1/hadiths/random":                       "Unable to load a random hadith at the moment.",
	"GET /v1/prophets":                             "Unable to load prophets at the moment.",
	"GET /v1/quran/verses":                         "Unable to load Quran verses at the moment.",
	"POST /v1/rewards":                             "Unable to create reward at the moment.",
	"PATCH /v1/rewards/:id":                        "Unable to update reward at the moment.",
	"DELETE /v1/rewards/:id":                       "Unable to delete reward at the moment.",
	"GET /v1/tasks":                                "Unable to load tasks at the moment.",
	"GET /v1/tasks/due-rewards":                    "Unable to load due rewards at the moment.",
	"GET /v1/tasks/:id":                            "Unable to load task details at the moment.",
	"GET /v1/tasks/recurring":                      "Unable to load recurring tasks at the moment.",
	"POST /v1/tasks/recurring":                     "Unable to create recurring task at the moment.",
	"DELETE /v1/tasks/recurring/:id":               "Unable to delete recurring task at the moment.",
	"POST /v1/quizzes/hadith":                      "Unable to assign hadith quiz at the moment.",
	"POST /v1/quizzes/prophet":                     "Unable to assign prophet quiz at the moment.",
	"POST /v1/quizzes/quran":                       "Unable to assign Quran quiz at the moment.",
	"POST /v1/quizzes/topic":                       "Unable to assign topic quiz at the moment.",
	"POST /v1/ai/ask":                              "Unable to process AI question at the moment.",
	"POST /v1/dua/generate":                        "Unable to generate dua at the moment.",
	"GET /v1/dua/history":                          "Unable to load dua history at the moment.",
	"GET /v1/dua/history/:id":                      "Unable to load the selected dua at the moment.",
	"POST /v1/quizzes/:type/:id/submit":            "Unable to submit quiz at the moment.",
	"GET /v1/lessons/quran":                        "Unable to load Quran lessons at the moment.",
	"POST /v1/lessons/quran":                       "Unable to create Quran lesson at the moment.",
	"POST /v1/lessons/quran/:id/complete":          "Unable to complete lesson at the moment.",
	"GET /v1/learn":                                "Unable to load learning content at the moment.",
	"POST /v1/learn":                               "Unable to create learning content at the moment.",
	"POST /v1/learn/:id/complete":                  "Unable to mark learning content complete at the moment.",
	"GET /v1/messages/conversations":               "Unable to load conversations at the moment.",
	"GET /v1/messages/:user_id":                    "Unable to load messages at the moment.",
	"POST /v1/messages":                            "Unable to send message at the moment.",
	"PATCH /v1/messages/:id/read":                  "Unable to mark message as read at the moment.",
	"GET /v1/requests":                             "Unable to load requests at the moment.",
	"POST /v1/requests":                            "Unable to create request at the moment.",
	"POST /v1/requests/:id/respond":                "Unable to respond to request at the moment.",
	"POST /v1/games/sessions/start":                "Unable to start game session at the moment.",
	"POST /v1/games/sessions/:id/end":              "Unable to end game session at the moment.",
	"GET /v1/dashboard/summary":                    "Unable to load dashboard summary at the moment.",
	"GET /v1/dashboard/task-completion":            "Unable to load task completion metrics at the moment.",
	"GET /v1/dashboard/game-time":                  "Unable to load game time metrics at the moment.",
	"GET /v1/dashboard/quiz-scores":                "Unable to load quiz score metrics at the moment.",
	"GET /v1/dashboard/learn-progress":             "Unable to load learning progress metrics at the moment.",
	"GET /v1/notifications":                        "Unable to load notifications at the moment.",
	"PATCH /v1/notifications/read-all":             "Unable to mark notifications as read at the moment.",
	"PATCH /v1/notifications/:id/read":             "Unable to mark notification as read at the moment.",
	"POST /v1/push/subscribe":                      "Unable to save push subscription at the moment.",
	"DELETE /v1/push/subscribe":                    "Unable to remove push subscription at the moment.",
	"GET /v1/rants":                                "Unable to load rants at the moment.",
	"POST /v1/rants":                               "Unable to create rant at the moment.",
	"GET /v1/rants/:id":                            "Unable to load rant at the moment.",
	"PATCH /v1/rants/:id":                          "Unable to update rant at the moment.",
	"DELETE /v1/rants/:id":                         "Unable to delete rant at the moment.",
}
