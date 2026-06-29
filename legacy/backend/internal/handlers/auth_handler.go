package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"modelmarket/internal/middleware"
	"modelmarket/internal/models"
	"modelmarket/internal/utils"
	"modelmarket/pkg/logger"
)

type AuthHandler struct {
	db *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler { return &AuthHandler{db: db} }

type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var u models.AdminUser
	if err := h.db.Where("username = ?", req.Username).First(&u).Error; err != nil {
		logger.L().Warnf("Login failed: user not found %s", req.Username)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if !utils.CheckPassword(u.PasswordHash, req.Password) {
		logger.L().Warnf("Login failed: bad password for %s", req.Username)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	tok, err := middleware.IssueToken(u.ID, u.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.L().Infof("Admin login ok: %s", u.Username)
	c.JSON(http.StatusOK, gin.H{"token": tok, "username": u.Username})
}

// Me GET /api/v1/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"uid":      c.GetUint("uid"),
		"username": c.GetString("username"),
	})
}
