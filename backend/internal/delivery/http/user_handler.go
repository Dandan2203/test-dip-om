package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"finagent/backend/internal/delivery/http/middleware"
	"finagent/backend/internal/domain"
)

type UserHandler struct {
	uc domain.UserUsecase
}

func NewUserHandler(uc domain.UserUsecase) *UserHandler {
	return &UserHandler{uc: uc}
}

type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
	Name     string `json:"name" binding:"max=255"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type updateProfileRequest struct {
	Name string `json:"name" binding:"max=255"`
}

type userResponse struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// toUserResponse — подання користувача без хеша пароля.
func toUserResponse(u *domain.User) userResponse {
	return userResponse{ID: u.ID, Email: u.Email, Name: u.Name}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	user, err := h.uc.Register(c.Request.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		HandleError(c, err)
		return
	}
	RespondOK(c, http.StatusCreated, toUserResponse(user))
}

func (h *UserHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	token, err := h.uc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		HandleError(c, err)
		return
	}
	RespondOK(c, http.StatusOK, gin.H{"token": token})
}

func (h *UserHandler) GetMe(c *gin.Context) {
	user, err := h.uc.GetProfile(c.Request.Context(), middleware.UserIDFromContext(c))
	if err != nil {
		HandleError(c, err)
		return
	}
	RespondOK(c, http.StatusOK, toUserResponse(user))
}

func (h *UserHandler) UpdateMe(c *gin.Context) {
	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	user, err := h.uc.UpdateProfile(c.Request.Context(), middleware.UserIDFromContext(c), req.Name)
	if err != nil {
		HandleError(c, err)
		return
	}
	RespondOK(c, http.StatusOK, toUserResponse(user))
}
