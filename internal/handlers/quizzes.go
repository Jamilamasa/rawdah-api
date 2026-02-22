package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/services"
)

type QuizHandler struct {
	svc *services.QuizService
}

func NewQuizHandler(svc *services.QuizService) *QuizHandler {
	return &QuizHandler{svc: svc}
}

func (h *QuizHandler) AssignHadith(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	assignedBy := c.GetString(string(models.ContextKeyUserID))

	var req struct {
		HadithID      string     `json:"hadith_id"      binding:"required"`
		AssignedTo    string     `json:"assigned_to"    binding:"required"`
		ChildAge      int        `json:"child_age"`
		MemorizeUntil *time.Time `json:"memorize_until"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	fid, err := uuid.Parse(familyID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	hid, err := uuid.Parse(req.HadithID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hadith_id"})
		return
	}
	aid, err := uuid.Parse(req.AssignedTo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assigned_to"})
		return
	}
	abid, err := uuid.Parse(assignedBy)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	quiz, err := h.svc.AssignHadith(c.Request.Context(), services.AssignHadithInput{
		FamilyID:      fid,
		HadithID:      hid,
		AssignedTo:    aid,
		AssignedBy:    abid,
		ChildAge:      req.ChildAge,
		MemorizeUntil: req.MemorizeUntil,
	})
	if err != nil {
		if err == services.ErrInvalidQuizAssignee {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, quiz)
}

func (h *QuizHandler) AssignProphet(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	assignedBy := c.GetString(string(models.ContextKeyUserID))

	var req struct {
		ProphetID  string `json:"prophet_id"  binding:"required"`
		AssignedTo string `json:"assigned_to" binding:"required"`
		ChildAge   int    `json:"child_age"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	fid, err := uuid.Parse(familyID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	pid, err := uuid.Parse(req.ProphetID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid prophet_id"})
		return
	}
	aid, err := uuid.Parse(req.AssignedTo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assigned_to"})
		return
	}
	abid, err := uuid.Parse(assignedBy)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	quiz, err := h.svc.AssignProphet(c.Request.Context(), services.AssignProphetInput{
		FamilyID:   fid,
		ProphetID:  pid,
		AssignedTo: aid,
		AssignedBy: abid,
		ChildAge:   req.ChildAge,
	})
	if err != nil {
		if err == services.ErrInvalidQuizAssignee {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, quiz)
}

func (h *QuizHandler) AssignQuran(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	assignedBy := c.GetString(string(models.ContextKeyUserID))

	var req struct {
		VerseID    string  `json:"verse_id"    binding:"required"`
		LessonID   *string `json:"lesson_id"`
		AssignedTo string  `json:"assigned_to" binding:"required"`
		ChildAge   int     `json:"child_age"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	fid, err := uuid.Parse(familyID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	vid, err := uuid.Parse(req.VerseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid verse_id"})
		return
	}
	aid, err := uuid.Parse(req.AssignedTo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assigned_to"})
		return
	}
	abid, err := uuid.Parse(assignedBy)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var lessonID *uuid.UUID
	if req.LessonID != nil {
		lid, err := uuid.Parse(*req.LessonID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson_id"})
			return
		}
		lessonID = &lid
	}

	quiz, err := h.svc.AssignQuran(c.Request.Context(), services.AssignQuranInput{
		FamilyID:   fid,
		VerseID:    vid,
		LessonID:   lessonID,
		AssignedTo: aid,
		AssignedBy: abid,
		ChildAge:   req.ChildAge,
	})
	if err != nil {
		if err == services.ErrInvalidQuizAssignee {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, quiz)
}

func (h *QuizHandler) List(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	role := c.GetString(string(models.ContextKeyRole))
	if role == "child" {
		h.ListMine(c)
		return
	}

	quizType := c.Query("type")

	type QuizListResponse struct {
		HadithQuizzes  interface{} `json:"hadith_quizzes,omitempty"`
		ProphetQuizzes interface{} `json:"prophet_quizzes,omitempty"`
		QuranQuizzes   interface{} `json:"quran_quizzes,omitempty"`
	}

	resp := QuizListResponse{}

	if quizType == "" || quizType == "hadith" {
		quizzes, err := h.svc.ListHadithQuizzes(c.Request.Context(), familyID)
		if err == nil {
			resp.HadithQuizzes = quizzes
		}
	}
	if quizType == "" || quizType == "prophet" {
		quizzes, err := h.svc.ListProphetQuizzes(c.Request.Context(), familyID)
		if err == nil {
			resp.ProphetQuizzes = quizzes
		}
	}
	if quizType == "" || quizType == "quran" {
		quizzes, err := h.svc.ListQuranQuizzes(c.Request.Context(), familyID)
		if err == nil {
			resp.QuranQuizzes = quizzes
		}
	}

	c.JSON(http.StatusOK, resp)
}

func (h *QuizHandler) ListMine(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	userID := c.GetString(string(models.ContextKeyUserID))

	type MyQuizzes struct {
		HadithQuizzes  interface{} `json:"hadith_quizzes"`
		ProphetQuizzes interface{} `json:"prophet_quizzes"`
		QuranQuizzes   interface{} `json:"quran_quizzes"`
	}

	resp := MyQuizzes{}

	hq, _ := h.svc.ListMyHadithQuizzes(c.Request.Context(), userID, familyID)
	pq, _ := h.svc.ListMyProphetQuizzes(c.Request.Context(), userID, familyID)
	qq, _ := h.svc.ListMyQuranQuizzes(c.Request.Context(), userID, familyID)

	resp.HadithQuizzes = hq
	resp.ProphetQuizzes = pq
	resp.QuranQuizzes = qq

	c.JSON(http.StatusOK, resp)
}

func (h *QuizHandler) Get(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	role := c.GetString(string(models.ContextKeyRole))
	userID := c.GetString(string(models.ContextKeyUserID))
	quizType := c.Param("type")
	id := c.Param("id")

	switch quizType {
	case "hadith":
		quiz, err := h.svc.GetHadithQuiz(c.Request.Context(), id, familyID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if role == "child" && quiz.AssignedTo.String() != userID {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, quiz)
	case "prophet":
		quiz, err := h.svc.GetProphetQuiz(c.Request.Context(), id, familyID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if role == "child" && quiz.AssignedTo.String() != userID {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, quiz)
	case "quran":
		quiz, err := h.svc.GetQuranQuiz(c.Request.Context(), id, familyID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if role == "child" && quiz.AssignedTo.String() != userID {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, quiz)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid quiz type"})
	}
}

func (h *QuizHandler) Start(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	userID := c.GetString(string(models.ContextKeyUserID))
	quizType := c.Param("type")
	id := c.Param("id")

	var err error
	switch quizType {
	case "hadith":
		err = h.svc.StartHadithQuiz(c.Request.Context(), id, familyID, userID)
	case "prophet":
		err = h.svc.StartProphetQuiz(c.Request.Context(), id, familyID, userID)
	case "quran":
		err = h.svc.StartQuranQuiz(c.Request.Context(), id, familyID, userID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid quiz type"})
		return
	}

	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "unable to start quiz"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "quiz started"})
}

func (h *QuizHandler) Submit(c *gin.Context) {
	familyID := c.GetString(string(models.ContextKeyFamilyID))
	userID := c.GetString(string(models.ContextKeyUserID))
	quizType := c.Param("type")
	id := c.Param("id")

	var req struct {
		Answers []models.QuizAnswer `json:"answers" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	input := services.SubmitAnswersInput{
		Answers: req.Answers,
		UserID:  userID,
	}

	switch quizType {
	case "hadith":
		result, err := h.svc.SubmitHadithQuiz(c.Request.Context(), id, familyID, input)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "unable to submit quiz"})
			return
		}
		c.JSON(http.StatusOK, result)
	case "prophet":
		result, err := h.svc.SubmitProphetQuiz(c.Request.Context(), id, familyID, input)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "unable to submit quiz"})
			return
		}
		c.JSON(http.StatusOK, result)
	case "quran":
		result, err := h.svc.SubmitQuranQuiz(c.Request.Context(), id, familyID, input)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "unable to submit quiz"})
			return
		}
		c.JSON(http.StatusOK, result)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid quiz type"})
	}
}
