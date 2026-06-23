package api

import (
	"errors"
	"net/http"

	appmw "github.com/alexscott64/woulder/backend/internal/api/middleware"
	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/alexscott64/woulder/backend/internal/service"
	"github.com/gin-gonic/gin"
)

func (h *Handler) AuthLogin(c *gin.Context) {
	var req models.AuthLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	resp, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		status := http.StatusUnauthorized
		if errors.Is(err, service.ErrAuthInactiveUser) {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": "Invalid credentials"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) AuthRefresh(c *gin.Context) {
	var req models.AuthRefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token is required"})
		return
	}
	resp, err := h.authService.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) AuthLogout(c *gin.Context) {
	var req models.AuthLogoutRequest
	_ = c.ShouldBindJSON(&req)
	if err := h.authService.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) AuthMe(c *gin.Context) {
	userID, _ := c.Get(appmw.ContextUserID)
	user, err := h.authService.CurrentUser(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}
