package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"synaply/external/gemini"
	"synaply/internal/middleware"
	"synaply/internal/models"
	"synaply/internal/repository/modelsDB"
	"synaply/internal/service"
	"synaply/internal/utils"
	"synaply/slogger"
)

const ctxWithTimeout time.Duration = time.Second * 5

type Handler struct {
	wordsService service.WordsService
	userService  service.UserService
}

func NewHandler(ws service.WordsService, us service.UserService) *Handler {
	return &Handler{wordsService: ws, userService: us}
}

// Lang Get available languages
// @Summary Get available languages
// @Description Returns a list of supported source and target languages for learning and translation
// @Tags users
// @Produce json
// @Success 200 {object} models.LangResponse "Successfully retrieved languages"
// @Router /users/lang [get]
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *Handler) Lang(w http.ResponseWriter, r *http.Request) {

	var Response models.LangResponse
	Response.Source = models.SortedSourceLanguages
	Response.Target = models.SortedTargetLanguages
	slogger.Log.Debug("lang called", "response", Response)
	JSONResponse(w, http.StatusOK, Response)
}

// Login
// @Summary User Login
// @Description Login and get JWT token
// @Tags users
// @Accept json
// @Produce json
// @Param input body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.LoginResponse
// @Failure 401 {object} handler.JSONError "Invalid credentials"
// @Router /login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		slogger.Log.DebugContext(r.Context(), "Invalid request body in Login", "err", err)
		return
	}
	if req.Email == "" || req.Password == "" {
		WriteError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	ctx := r.Context()
	slogger.Log.DebugContext(ctx, "Login request", "req", req)

	token, sourceLang, err := h.userService.Login(ctx, req)
	slogger.Log.DebugContext(ctx, "Login response", "token", token)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			WriteError(w, http.StatusUnauthorized, "Invalid email or password")
			return
		}
		WriteError(w, http.StatusInternalServerError, "Internal server error")
		slogger.Log.ErrorContext(ctx, "Login failed", "err", err)
		return
	}

	JSONResponse(w, http.StatusOK, models.LoginResponse{Token: token, SourceLang: sourceLang})
}

// Create creates a new user
// @Summary Create a new user
// @Description Register a new user in the system
// @Tags users
// @Accept json
// @Produce json
// @Param input body models.CreateUserRequest true "User registration info"
// @Success 201 {object} models.UserResponse
// @Failure 400 {object} handler.JSONError "Invalid input or validation info"
// @Failure 409 {object} handler.JSONError "User already exists"
// @Failure 500 {object} handler.JSONError "Internal server error"
// @Router /users [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {

	var req models.CreateUserRequest

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		slogger.Log.ErrorContext(r.Context(), "Invalid request body", "err", err)
		return
	}

	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" || req.SourceLang == "" || req.TargetLang == "" {
		WriteError(w, http.StatusBadRequest, "Fields cannot be empty")
		slogger.Log.DebugContext(r.Context(), "Registration failed: empty fields", "req", req)
		return
	}
	if req.SourceLang == req.TargetLang {
		WriteError(w, http.StatusBadRequest, "Target lang cannot be same as source lang")
		slogger.Log.DebugContext(r.Context(), "Registration failed: target lang cannot be same as source lang", "req", req)
		return
	}
	if err := utils.ValidatePassword(req.Password); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		slogger.Log.DebugContext(r.Context(), "Registration failed: invalid password", "err", err)
		return
	}

	ctx := r.Context()

	createdUser, err := h.userService.Create(ctx, req)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrUserAlreadyExists):
			WriteError(w, http.StatusConflict, "User already exists")
			return
		}
		WriteError(w, http.StatusInternalServerError, err.Error())
		slogger.Log.ErrorContext(ctx, "Failed to create user", "err", err)
		return
	}

	JSONResponse(w, http.StatusCreated, createdUser)
}

// GetUserByID Get User By Slug
// @Summary Get user by Slug
// @Description Get User By Slug from DB
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User UUID" format(uuid)
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} handler.JSONError "Invalid UUID"
// @Failure 404 {object} handler.JSONError "User not found"
// @Failure 500 {object} handler.JSONError "Internal server error"
// @Router /users/{id} [get]
func (h *Handler) GetUserByID(w http.ResponseWriter, r *http.Request) {

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid user Slug")
		slogger.Log.DebugContext(r.Context(), "Invalid user Slug", "err", err, "id", chi.URLParam(r, "id"))
		return
	}

	ctx := r.Context()

	user, err := h.userService.GetUserByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrUserNotFound):
			WriteError(w, http.StatusNotFound, "User not found")
			slogger.Log.DebugContext(ctx, "User not found", "err", err, "id", id)
			return
		default:
			WriteError(w, http.StatusInternalServerError, "Failed to get user")
			slogger.Log.ErrorContext(ctx, "Failed to get user", "err", err)
			return
		}
	}
	JSONResponse(w, http.StatusOK, user)
}

// GetUsers Get list of users
// @Summary Get list of users
// @Description Get list of all users from DB
// @Tags users
// @Accept json
// @Produce json
// @Param limit query int false "Limit records per page (default 10)"
// @Param page query int false "Page number (default 1)"
// @Param order query string false "Sort order: asc or desc (default desc)"
// @Success 200 {object} models.ListOfUsersResponse
// @Failure 500 {object} handler.JSONError "Failed to get users"
// @Router /users [get]
func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {

	var (
		limit uint64 = 10
		page  uint64 = 1
	)

	if s := r.URL.Query().Get("limit"); s != "" {
		l, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "Invalid limit")
			return
		}
		limit = l
	}

	if s := r.URL.Query().Get("page"); s != "" {
		p, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "Invalid page")
			return
		}
		page = p
	}

	order := r.URL.Query().Get("order")
	if order != "asc" {
		order = "desc"
	}

	ctx := r.Context()

	users, err := h.userService.GetUsers(ctx, limit, page, order)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to get users")
		slogger.Log.ErrorContext(ctx, "Failed to get users", "err", err)
		return
	}
	JSONResponse(w, http.StatusOK, users)
}

// Delete User delete by Slug
// @Summary Delete user by Slug
// @Description Delete user by Slug from DB
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User UUID" format(uuid)
// @Success 200 {object} nil
// @Failure 400 {object} handler.JSONError "Invalid user Slug"
// @Failure 403 {object} handler.JSONError "You can delete only your own account"
// @Failure 500 {object} handler.JSONError "Failed to delete user"
// @Router /users/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	targetID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid user Slug")
		slogger.Log.DebugContext(r.Context(), "Invalid user Slug in Delete", "err", err, "id", chi.URLParam(r, "id"))
		return
	}
	target, err := h.userService.GetUserByID(r.Context(), targetID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to get user")
		slogger.Log.ErrorContext(r.Context(), "Failed to get user in Delete", "err", err, "target_id", targetID)
		return
	}
	user, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		slogger.Log.ErrorContext(r.Context(), "Unauthorized", "error", err)
		return
	}

	ctx := r.Context()

	if !models.PermissionCheck(user, target) {
		WriteError(w, http.StatusForbidden, models.ErrPermissionDenied.Error())
		return
	}
	err = h.userService.Delete(ctx, targetID)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrUserNotFound):
			WriteError(w, http.StatusNotFound, "User not found")
			return
		case errors.Is(err, models.ErrPermissionDenied):
			WriteError(w, http.StatusForbidden, models.ErrPermissionDenied.Error())
			return

		}
		WriteError(w, http.StatusInternalServerError, "Failed to delete user")
		slogger.Log.ErrorContext(ctx, "Failed to delete user", "err", err)
		return
	}

	JSONResponse(w, http.StatusOK, nil)
}

// Update User update by Slug
// @Summary Update user by Slug
// @Description Update user information by its UUID
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User UUID" format(uuid)
// @Param input body models.UpdateUserRequest true "User update info"
// @Success 200 {object} models.UserResponse "Successfully updated"
// @Failure 400 {object} handler.JSONError "Invalid user Slug"
// @Failure 403 {object} handler.JSONError "Permission denied"
// @Failure 500 {object} handler.JSONError "Failed to update user"
// @Router /users/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {

	targetID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid user Slug")
		slogger.Log.DebugContext(r.Context(), "Invalid user Slug in Update", "err", err, "id", chi.URLParam(r, "id"))
		return
	}

	user, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		slogger.Log.ErrorContext(r.Context(), "Unauthorized", "error", err)
		return
	}
	slogger.Log.DebugContext(r.Context(), "Updating user", "targetID", targetID, "requester", user)

	target, err := h.userService.GetUserByID(r.Context(), targetID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to get user")
		slogger.Log.ErrorContext(r.Context(), "Failed to get user in Update", "err", err, "target_id", targetID)
		return
	}

	if !models.PermissionCheck(user, target) {
		WriteError(w, http.StatusForbidden, models.ErrPermissionDenied.Error())
		return
	}

	var req models.UpdateUserRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		slogger.Log.ErrorContext(r.Context(), "Invalid request body", "err", err)
		return
	}
	ctx := r.Context()

	if req.Role != nil && user.Role != models.RoleAdmin {
		WriteError(w, http.StatusBadRequest, models.ErrPermissionDenied.Error())
		return
	}
	if req.Password != nil {
		if err := utils.ValidatePassword(*req.Password); err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	if req.SourceLang != nil && req.TargetLang != nil && *req.SourceLang == *req.TargetLang {
		WriteError(w, http.StatusBadRequest, "Target lang cannot be same as source lang")
		return
	}

	updatedUser, err := h.userService.Update(ctx, targetID, req)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrUserNotFound):
			WriteError(w, http.StatusNotFound, "User not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "Failed to update user")
		slogger.Log.ErrorContext(ctx, "Failed to update user", "err", err)
		return
	}

	JSONResponse(w, http.StatusOK, updatedUser)
}

// GetMe Get current user info
// @Summary Get current user info
// @Description Returns language settings and total correct answers for the current user
// @Tags words
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.GetMeResponse
// @Failure 401 {object} handler.JSONError "Unauthorized"
// @Failure 500 {object} handler.JSONError "Internal server error"
// @Router /words/GetMe [get]
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		slogger.Log.ErrorContext(ctx, "Unauthorized", "error", err)
		return
	}

	user, err := h.userService.GetUserByID(ctx, userCtx.ID)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			WriteError(w, http.StatusUnauthorized, "User not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "Failed to get user")
		slogger.Log.ErrorContext(ctx, "Failed to get user", "err", err)
		return
	}
	var langCodeResp models.LangCodeResp
	langCodeResp.Source = user.SourceLang
	langCodeResp.Target = user.TargetLang

	response := models.GetMeResponse{
		ID:           user.ID,
		Email:        user.Email,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Role:         user.Role,
		SourceLang:   user.SourceLang,
		LangCodeResp: langCodeResp,
		TotalCorrect: user.TotalCorrect,
	}

	JSONResponse(w, http.StatusOK, response)
}

// Translate Translate a word
// @Summary Translate a word
// @Description Translates a word using Gemini API
// @Tags words
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body models.TranslateReq true "Translation request data"
// @Success 200 {object} models.TranslateResp
// @Failure 400 {object} handler.JSONError "Invalid request body"
// @Failure 401 {object} handler.JSONError "Unauthorized"
// @Failure 500 {object} handler.JSONError "Internal server error"
// @Router /words/translate [post]
func (h *Handler) Translate(w http.ResponseWriter, r *http.Request) {
	slogger.Log.DebugContext(r.Context(), "Translate")
	var req models.TranslateReq
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		slogger.Log.DebugContext(r.Context(), "Invalid request body in Translate", "err", err)
		return
	}
	if strings.TrimSpace(req.SourceWord) == "" && strings.TrimSpace(req.TargetWord) == "" {
		WriteError(w, http.StatusBadRequest, "Both source and target words are empty")
		return
	}

	ctx := r.Context()

	userCtx, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		slogger.Log.ErrorContext(ctx, "Unauthorized", "error", err)
		return
	}

	user, err := h.userService.GetUserByID(ctx, userCtx.ID)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			WriteError(w, http.StatusUnauthorized, "User not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "Failed to get user")
		slogger.Log.ErrorContext(ctx, "Failed to get user", "err", err)
		return
	}
	//todo
	req.ID = userCtx.ID
	req.SourceLang = user.SourceLang
	req.TargetLang = user.TargetLang
	slogger.Log.Debug("translate req %+v", "req", req)
	resp, err := h.wordsService.Translate(ctx, req)
	if err != nil {
		if errors.Is(err, gemini.ErrLimitExceeded) {
			WriteError(w, http.StatusTooManyRequests, "Sorry, but the free mode has ended")
			return
		}
		WriteError(w, http.StatusInternalServerError, "Failed to translate")
		slogger.Log.ErrorContext(ctx, "Translation failed", "error", err)
		return
	}
	slogger.Log.Debug("translate resp %+v", "resp", resp)

	JSONResponse(w, http.StatusOK, resp)
}

// NewWord Create a new word
// @Summary Create a new word
// @Description Adds a new word to the user's dictionary
// @Tags words
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body models.CreateReq true "Word data"
// @Success 200 {object} models.Response
// @Failure 400 {object} handler.JSONError "Invalid request body"
// @Failure 401 {object} handler.JSONError "Unauthorized"
// @Failure 408 {object} handler.JSONError "Database timeout"
// @Failure 409 {object} handler.JSONError "Word already exists"
// @Failure 500 {object} handler.JSONError "Internal server error"
// @Router /words/create [post]
func (h *Handler) NewWord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req models.CreateReq

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		slogger.Log.ErrorContext(ctx, "Invalid request body", "error", err)
		return
	}
	slogger.Log.DebugContext(ctx, "Handler NewWord called with req", "req", req)

	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		slogger.Log.ErrorContext(ctx, "Unauthorized", "error", err)
		return
	}

	resp, err := h.wordsService.Create(ctx, req, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrWordAlreadyExists):
			WriteError(w, http.StatusConflict, "Word already exists")
			slogger.Log.ErrorContext(ctx, "Word already exists", "error", err)
			return
		case errors.Is(err, models.ErrDBTimeout):
			WriteError(w, http.StatusRequestTimeout, "Timeout")
			slogger.Log.ErrorContext(ctx, "Timeout", "error", err)
			return
		default:
			WriteError(w, http.StatusInternalServerError, "Internal server error")
			slogger.Log.ErrorContext(ctx, "Internal server error", "error", err)
			return
		}
	}

	slogger.Log.DebugContext(ctx, "Handler NewWord called with resp", "resp", resp)
	JSONResponse(w, http.StatusOK, resp)
}

// StartLesson Start a new lesson
// @Summary Start a learning lesson
// @Description Gets the next word for the user to learn
// @Tags lesson
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.StartLessonResponse
// @Failure 401 {object} handler.JSONError "Unauthorized"
// @Failure 404 {object} handler.JSONError "no_words_for_lesson"
// @Failure 500 {object} handler.JSONError "Internal server error"
// @Router /lesson/start [get]
func (h *Handler) StartLesson(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		slogger.Log.ErrorContext(ctx, "Unauthorized", "error", err)
		return
	}

	word, err := h.wordsService.LessonStart(ctx, userCtx.ID)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNoWordsForLesson):
			WriteError(w, http.StatusNotFound, "no_words_for_lesson")
			slogger.Log.ErrorContext(ctx, "No words found for lesson", "error", err)
			return
		default:
			WriteError(w, http.StatusInternalServerError, "Internal server error")
			slogger.Log.ErrorContext(ctx, "Internal server error", "error", err)
			return
		}
	}

	totalCorrect, err := h.userService.GetTotalCorrect(ctx, userCtx.ID)
	if err != nil {
		slogger.Log.ErrorContext(ctx, "Failed to get total correct count", "error", err)
	}

	response := models.StartLessonResponse{
		TotalCorrect: totalCorrect,
		Word:         word,
	}

	JSONResponse(w, http.StatusOK, response)

}

// CheckAnswer Check word answer
// @Summary Check the answer for a word
// @Description Submits an answer for the current word and returns the next word
// @Tags lesson
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body models.AnswerReq true "Answer data"
// @Success 200 {object} models.CheckAnswerResponse
// @Failure 400 {object} handler.JSONError "Invalid request body"
// @Failure 401 {object} handler.JSONError "Unauthorized"
// @Failure 500 {object} handler.JSONError "Internal server error"
// @Router /lesson/check [post]
func (h *Handler) CheckAnswer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		slogger.Log.ErrorContext(ctx, "Unauthorized", "error", err)
		return
	}

	var req models.AnswerReq
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		slogger.Log.ErrorContext(ctx, "Invalid request body", "error", err)
		return
	}
	slogger.Log.DebugContext(ctx, "Handler CheckAnswer called with req", "req", req)

	if req.ID == "" {
		slogger.Log.ErrorContext(ctx, "Handler CheckAnswer called with empty req.Slug", "req", req.ID)
		WriteError(w, http.StatusBadRequest, "Invalid request Slug")
		return
	}
	isCorrect, resp, err := h.wordsService.CheckAnswer(ctx, req, user.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to check answer")
		slogger.Log.ErrorContext(ctx, "Failed to check answer", "err", err)
		return
	}

	totalCorrect, err := h.userService.GetTotalCorrect(ctx, user.ID)
	if err != nil {
		//todo
	}
	var totalCorrectUpdate int64
	if isCorrect {
		totalCorrectUpdate, err = h.userService.SetTotalCorrect(ctx, user.ID, totalCorrect)
		if err != nil {

		}
	} else {
		totalCorrectUpdate = totalCorrect
	}

	response := models.CheckAnswerResponse{
		IsCorrect:    isCorrect,
		TotalCorrect: totalCorrectUpdate,
		NextWord:     resp,
	}

	if isCorrect {
		slogger.Log.DebugContext(ctx, "Handler CheckAnswer is correct")
	} else {
		slogger.Log.DebugContext(ctx, "Handler CheckAnswer is NOT correct")
	}

	JSONResponse(w, http.StatusOK, response)

}

// Finish Finish current lesson
// @Summary Finish the learning lesson
// @Description Clears the current lesson state from the cache for the current user
// @Tags lesson
// @Produce json
// @Security BearerAuth
// @Success 200 {string} string "OK"
// @Failure 401 {object} handler.JSONError "Unauthorized"
// @Router /lesson/finish [post]
func (h *Handler) Finish(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		slogger.Log.ErrorContext(ctx, "Unauthorized", "error", err)
		return
	}

	h.wordsService.Finish(ctx, user.ID)
	JSONResponse(w, http.StatusOK, nil)
}

// WordInfo Get AI-generated information about a word
// @Summary Get word details using AI
// @Description Returns detailed explanation, grammar, and usage examples for a word using Gemini AI
// @Tags words
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body gemini.WordInfoRequest true "Data for Gemini AI prompt"
// @Success 200 {object} gemini.WordInfoResponse "Successfully generated explanation"
// @Failure 400 {object} handler.JSONError "Invalid request body"
// @Failure 401 {object} handler.JSONError "Unauthorized"
// @Failure 500 {object} handler.JSONError "Internal server error"
// @Router /words/wordInfo [post]
func (h *Handler) WordInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		slogger.Log.ErrorContext(ctx, "Unauthorized", "error", err)
		return
	}

	var req gemini.WordInfoRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		slogger.Log.DebugContext(ctx, "Invalid request body", "err", err)
		return
	}

	user, err := h.userService.GetUserByID(ctx, userCtx.ID)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			WriteError(w, http.StatusUnauthorized, "User not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "Failed to get user")
		slogger.Log.ErrorContext(ctx, "Failed to get user", "err", err)
		return
	}

	req.SourceLang = user.SourceLang
	req.TargetLang = user.TargetLang

	resp, err := h.wordsService.WordInfo(ctx, req)
	if err != nil {
		if errors.Is(err, gemini.ErrLimitExceeded) {
			WriteError(w, http.StatusTooManyRequests, "Sorry, but the free mode has ended")
			return
		}
		WriteError(w, http.StatusInternalServerError, "Failed to get word info")
		slogger.Log.ErrorContext(ctx, "Failed to get word info", "err", err)
		return
	}
	JSONResponse(w, http.StatusOK, resp)

}

// StartPracticeWithGemini Start a practice session with Gemini
// @Summary Start AI practice
// @Description Generates a set of contextual sentences for translation based on a topic using Gemini AI
// @Tags AI Practice
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body object true "Topic for practice"
// @Success 200 {object} gemini.StartPracticeWithGeminiResponse
// @Failure 400 {object} JSONError
// @Failure 401 {object} JSONError
// @Failure 429 {object} JSONError
// @Failure 500 {object} JSONError
// @Router /practice/startPractice [post]
func (h *Handler) StartPracticeWithGemini(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		slogger.Log.ErrorContext(ctx, "Unauthorized", "error", err)
		return
	}

	user, err := h.userService.GetUserByID(ctx, userCtx.ID)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			WriteError(w, http.StatusUnauthorized, "User not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "Failed to get user")
		slogger.Log.ErrorContext(ctx, "Failed to get user", "err", err)
		return
	}

	var reqBody struct {
		Topic string `json:"topic"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(reqBody.Topic) > 300 {
		WriteError(w, http.StatusBadRequest, "Topic must be less than 300 characters")
		return
	}

	gemReq := &gemini.PracticeWithGemini{
		SourceLang: user.SourceLang,
		TargetLang: user.TargetLang,
		Topic:      reqBody.Topic,
	}
	resp, err := h.wordsService.StartPracticeWithGemini(ctx, gemReq, userCtx.ID)
	if err != nil {
		if errors.Is(err, gemini.ErrLimitExceeded) {
			WriteError(w, http.StatusTooManyRequests, "Sorry, but the free mode has ended")
			return
		}
		WriteError(w, http.StatusInternalServerError, "Failed to start practice")
		slogger.Log.ErrorContext(ctx, "Failed to start practice", "err", err)
		return
	}
	JSONResponse(w, http.StatusOK, resp)
}

// CheckAnswerPracticeWithGemini Check answers for AI practice
// @Summary Check AI practice answer
// @Description Submits user translations for the generated sentences and returns AI feedback
// @Tags AI Practice
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body models.UserTranslateResponse true "User translations"
// @Success 200 {object} gemini.CheckAnswerPracticeWithGeminiResponse
// @Failure 400 {object} JSONError
// @Failure 401 {object} JSONError
// @Failure 429 {object} JSONError
// @Failure 500 {object} JSONError
// @Router /practice/checkAnswerPractice [post]
func (h *Handler) CheckAnswerPracticeWithGemini(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		slogger.Log.ErrorContext(ctx, "Unauthorized", "error", err)
		return
	}

	user, err := h.userService.GetUserByID(ctx, userCtx.ID)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			WriteError(w, http.StatusUnauthorized, "User not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "Failed to get user")
		slogger.Log.ErrorContext(ctx, "Failed to get user", "err", err)
		return
	}

	var userTranslateResponse models.UserTranslateResponse
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	if err := json.NewDecoder(r.Body).Decode(&userTranslateResponse); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	gemReq := &gemini.PracticeWithGemini{
		SourceLang: user.SourceLang,
		TargetLang: user.TargetLang,
		Topic:      "",
	}
	resp, err := h.wordsService.CheckAnswerPracticeWithGemini(ctx, gemReq, userCtx.ID, userTranslateResponse.Translation)
	if err != nil {
		if errors.Is(err, gemini.ErrLimitExceeded) {
			WriteError(w, http.StatusTooManyRequests, "Sorry, but the free mode has ended")
			return
		}
		WriteError(w, http.StatusInternalServerError, "Failed to check answer")
		slogger.Log.ErrorContext(ctx, "Failed to check answer", "err", err)
		return
	}
	JSONResponse(w, http.StatusOK, resp)
}

// FinishPracticeWithGemini Finish the AI practice session
// @Summary Finish AI practice
// @Description Clears the AI practice session from the cache
// @Tags AI Practice
// @Produce json
// @Security BearerAuth
// @Success 200 {string} string "ok"
// @Failure 401 {object} JSONError
// @Failure 500 {object} JSONError
// @Router /practice/finishPractice [post]
func (h *Handler) FinishPracticeWithGemini(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		slogger.Log.ErrorContext(ctx, "Unauthorized", "error", err)
		return
	}

	err = h.wordsService.FinishPracticeWithGemini(ctx, userCtx.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to finish practice")
		slogger.Log.ErrorContext(ctx, "Failed to finish practice", "err", err)
		return
	}
	JSONResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) GetWordsList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		slogger.Log.ErrorContext(ctx, "Unauthorized", "error", err)
		return
	}

	search := r.URL.Query().Get("search")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 30
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil && offsetStr != "" {
		slogger.Log.DebugContext(ctx, "Invalid offset in GetWordsList", "err", err, "offset", offsetStr)
	}

	req := modelsDB.GetWordsListReq{
		UserID: user.ID,
		Search: search,
		Limit:  limit,
		Offset: offset,
	}

	words, total, err := h.wordsService.GetWordsList(ctx, req)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Internal server error")
		slogger.Log.ErrorContext(ctx, "Failed to get words list", "err", err)
		return
	}

	JSONResponse(w, http.StatusOK, map[string]any{
		"words": words,
		"total": total,
	})
}

func (h *Handler) UpdateWordFields(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		slogger.Log.ErrorContext(ctx, "Unauthorized", "error", err)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid Slug")
		slogger.Log.DebugContext(ctx, "Invalid Slug in UpdateWordFields", "err", err, "id", idStr)
		return
	}

	var req modelsDB.UpdateWordReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		slogger.Log.ErrorContext(ctx, "Invalid request body", "error", err)
		return
	}
	req.ID = id

	err = h.wordsService.UpdateWordFields(ctx, req, user.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Internal server error")
		slogger.Log.ErrorContext(ctx, "Failed to update word fields", "err", err, "word_id", id)
		return
	}

	JSONResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) DeleteWord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		slogger.Log.ErrorContext(ctx, "Unauthorized", "error", err)
		return
	}

	idStr := chi.URLParam(r, "id")
	err = h.wordsService.DeleteWord(ctx, idStr, user.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Internal server error")
		slogger.Log.ErrorContext(ctx, "Failed to delete word", "err", err, "word_id", idStr)
		return
	}

	JSONResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) DeleteAllWords(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		slogger.Log.ErrorContext(ctx, "Unauthorized", "error", err)
		return
	}

	err = h.wordsService.DeleteAllWords(ctx, user.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Internal server error")
		slogger.Log.ErrorContext(ctx, "Failed to delete all words", "err", err, "user_id", user.ID)
		return
	}

	JSONResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

// WordList Generate a thematic word list using AI
// @Summary Generate word list
// @Description Returns a list of thematic words based on user request or predefined topic using Gemini AI
// @Tags AI Features
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body models.WordListReq true "Topic and request for word list generation"
// @Success 200 {array} models.WordListResp
// @Failure 400 {object} JSONError
// @Failure 401 {object} JSONError
// @Failure 429 {object} JSONError
// @Failure 500 {object} JSONError
// @Router /words/wordList [post]
func (h *Handler) WordList(w http.ResponseWriter, r *http.Request) {
	const wordListTimeout = 120 * time.Second
	ctx, cancel := context.WithTimeout(r.Context(), wordListTimeout)
	defer cancel()

	userCtx, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		slogger.Log.ErrorContext(ctx, "Unauthorized", "error", err)
		return
	}

	user, err := h.userService.GetUserByID(ctx, userCtx.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to get user")
		slogger.Log.ErrorContext(ctx, "Failed to get user", "error", err)
		return
	}

	var wordListReq models.WordListReq
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	if err := json.NewDecoder(r.Body).Decode(&wordListReq); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		slogger.Log.ErrorContext(ctx, "Invalid request body", "error", err)
		return
	}
	slogger.Log.DebugContext(ctx, "WordList request body decoded", "req", wordListReq)
	wordListReq.SourceLang = user.SourceLang
	wordListReq.TargetLang = user.TargetLang

	if wordListReq.Topic == "" && wordListReq.UserTopic == "" {
		WriteError(w, http.StatusBadRequest, "Please select a topic or provide a custom request")
		return
	}

	slogger.Log.DebugContext(ctx, "WordList request", "req", wordListReq)
	resp, err := h.wordsService.WordList(ctx, user, wordListReq)

	if err != nil {
		if errors.Is(err, gemini.ErrLimitExceeded) {
			WriteError(w, http.StatusTooManyRequests, "Sorry, but the free mode has ended. Please try again tomorrow.")
			return
		}
		// Corrected error message
		WriteError(w, http.StatusInternalServerError, "Failed to generate word list. Please try again with a different topic or request.")
		slogger.Log.ErrorContext(ctx, "Failed to generate word list", "error", err, "req", wordListReq)
		return
	}
	JSONResponse(w, http.StatusOK, resp)
}

// CreateBatchWords mass save words
// @Summary Batch save words
// @Description Adds multiple words (up to 500) to the user's dictionary in a single request
// @Tags words
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body models.CreateBatchReq true "Batch of words to save"
// @Success 200 {object} map[string]string
// @Failure 400 {object} JSONError
// @Failure 401 {object} JSONError
// @Failure 500 {object} JSONError
// @Router /words/create-batch [post]
func (h *Handler) CreateBatchWords(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.CreateBatchReq
	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // 1MB limit
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body format")
		slogger.Log.ErrorContext(ctx, "Invalid request body format in CreateBatchWords", "error", err)
		return
	}

	if len(req.Words) > 500 {
		WriteError(w, http.StatusBadRequest, "Too many words in one batch (maximum is 500)")
		return
	}

	err = h.wordsService.CreateBatch(ctx, req, user.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to save words")
		slogger.Log.ErrorContext(ctx, "Failed to CreateBatchWords", "error", err)
		return
	}

	JSONResponse(w, http.StatusOK, map[string]string{"status": "success", "message": "Words saved"})
}

// ImportWords Import words from CSV or JSON file
// @Summary Import words
// @Description Uploads and processes a CSV or JSON file to add multiple words to the user's dictionary
// @Tags words
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "CSV or JSON file with words"
// @Success 200 {object} map[string]string
// @Failure 400 {object} JSONError
// @Failure 401 {object} JSONError
// @Failure 500 {object} JSONError
// @Router /words/import [post]
func (h *Handler) ImportWords(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userCtx, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.userService.GetUserByID(ctx, userCtx.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	// Parse multipart form
	err = r.ParseMultipartForm(5 << 20) // 5MB limit
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Failed to parse form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "File is required")
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to read file")
		return
	}

	id, err := uuid.Parse(user.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to parse user ID")
		return
	}

	err = h.wordsService.ImportWords(ctx, fileBytes, header.Filename, id, user.SourceLang, user.TargetLang)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		slogger.Log.ErrorContext(ctx, "Failed to import words", "error", err)
		return
	}

	JSONResponse(w, http.StatusOK, map[string]string{"status": "success", "message": "Words imported successfully"})
}

func (h *Handler) GetProgressStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		slogger.Log.ErrorContext(ctx, "Unauthorized", "error", err)
		return
	}

	stats, err := h.wordsService.GetProgressStats(ctx, user.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Internal server error")
		slogger.Log.ErrorContext(ctx, "Failed to get progress stats", "err", err, "user_id", user.ID)
		return
	}

	JSONResponse(w, http.StatusOK, stats)
}

func (h *Handler) CreatePublicWordList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.CreatePublicWordListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Title == "" || len(req.Words) == 0 || req.Level == "" {
		WriteError(w, http.StatusBadRequest, "Title, level and words are required")
		return
	}

	id, err := h.wordsService.CreatePublicWordList(ctx, req, user.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to create public word list")
		slogger.Log.ErrorContext(ctx, "Failed to create public word list", "err", err)
		return
	}

	JSONResponse(w, http.StatusCreated, map[string]any{"id": id})
}

// GetPublicWordLists returns all public word lists
// @Summary Get public word lists
// @Description Returns a list of all public word lists, optionally filtered by language
// @Tags words
// @Produce json
// @Param source_lang query string false "Source language code (e.g., 'en')"
// @Param target_lang query string false "Target language code (e.g., 'ru')"
// @Param level query string false "Proficiency level (e.g., 'A1')"
// @Success 200 {array} modelsDB.PublicWordList
// @Router /public-lists [get]
func (h *Handler) GetPublicWordLists(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sourceLang := r.URL.Query().Get("source_lang")
	targetLang := r.URL.Query().Get("target_lang")
	level := r.URL.Query().Get("level")

	lists, err := h.wordsService.GetPublicWordLists(ctx, sourceLang, targetLang, level)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to get public word lists")
		slogger.Log.ErrorContext(ctx, "Failed to get public word lists", "err", err)
		return
	}

	JSONResponse(w, http.StatusOK, lists)
}

// GetPublicWordListByID returns a public word list with its items
// @Summary Get public word list by ID
// @Description Returns detailed information about a public word list, including all its words
// @Tags words
// @Produce json
// @Param id path string true "Word list ID"
// @Success 200 {object} modelsDB.PublicWordListDetail
// @Router /public-lists/{id} [get]
func (h *Handler) GetPublicWordListByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	detail, err := h.wordsService.GetPublicWordListByID(ctx, id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to get public word list")
		slogger.Log.ErrorContext(ctx, "Failed to get public word list", "err", err, "id", id)
		return
	}

	JSONResponse(w, http.StatusOK, detail)
}

func (h *Handler) UpdatePublicWordList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	var req models.CreatePublicWordListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Title == "" || len(req.Words) == 0 || req.Level == "" {
		WriteError(w, http.StatusBadRequest, "Title, level and words are required")
		return
	}

	err = h.wordsService.UpdatePublicWordList(ctx, id, req, user.ID)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			WriteError(w, http.StatusForbidden, err.Error())
			return
		}
		WriteError(w, http.StatusInternalServerError, "Failed to update public word list")
		slogger.Log.ErrorContext(ctx, "Failed to update public word list", "err", err, "id", id)
		return
	}

	JSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *Handler) AddPublicListToUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	err = h.wordsService.AddPublicListToUser(ctx, id, user.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to add words to dictionary")
		slogger.Log.ErrorContext(ctx, "Failed to add public list to user", "err", err, "list_id", id, "user_id", user.ID)
		return
	}

	JSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *Handler) CreatePlaylist(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.CreatePlaylistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Title == "" {
		WriteError(w, http.StatusBadRequest, "Title is required")
		return
	}

	id, err := h.wordsService.CreatePlaylist(ctx, req, user.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to create playlist")
		slogger.Log.ErrorContext(ctx, "Failed to create playlist", "err", err)
		return
	}

	JSONResponse(w, http.StatusCreated, map[string]any{"id": id})
}

// GetPlaylists returns all playlists
// @Summary Get playlists
// @Description Returns a list of all playlists (collections of public word lists)
// @Tags words
// @Produce json
// @Success 200 {array} modelsDB.Playlist
// @Router /playlists [get]
func (h *Handler) GetPlaylists(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	playlists, err := h.wordsService.GetPlaylists(ctx)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to get playlists")
		slogger.Log.ErrorContext(ctx, "Failed to get playlists", "err", err)
		return
	}

	JSONResponse(w, http.StatusOK, playlists)
}

// GetPlaylistByID returns a playlist with its public word lists
// @Summary Get playlist by ID
// @Description Returns detailed information about a playlist, including all its public word lists
// @Tags words
// @Produce json
// @Param id path string true "Playlist ID"
// @Success 200 {object} modelsDB.PlaylistDetail
// @Router /playlists/{id} [get]
func (h *Handler) GetPlaylistByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	detail, err := h.wordsService.GetPlaylistByID(ctx, id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to get playlist")
		slogger.Log.ErrorContext(ctx, "Failed to get playlist", "err", err, "id", id)
		return
	}

	JSONResponse(w, http.StatusOK, detail)
}

func (h *Handler) UpdatePlaylist(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	var req models.CreatePlaylistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err = h.wordsService.UpdatePlaylist(ctx, id, req, user.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to update playlist")
		slogger.Log.ErrorContext(ctx, "Failed to update playlist", "err", err, "id", id)
		return
	}

	JSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *Handler) DeletePlaylist(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	err = h.wordsService.DeletePlaylist(ctx, id, user.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to delete playlist")
		slogger.Log.ErrorContext(ctx, "Failed to delete playlist", "err", err, "id", id)
		return
	}

	JSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *Handler) GetAdminStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	search := r.URL.Query().Get("search")
	stats, err := h.userService.GetAdminStats(ctx, search)
	if err != nil {
		slogger.Log.ErrorContext(ctx, "Failed to get admin stats", "err", err)
		WriteError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	JSONResponse(w, http.StatusOK, stats)
}
