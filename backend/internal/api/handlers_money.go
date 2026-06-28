package api

import (
	"errors"
	"net/http"
	"strings"
	"time"

	appmw "github.com/alexscott64/woulder/backend/internal/api/middleware"
	"github.com/alexscott64/woulder/backend/internal/database/dberrors"
	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/alexscott64/woulder/backend/internal/service"
	"github.com/gin-gonic/gin"
)

func (h *Handler) GetMoneyProject(c *gin.Context) {
	resp, err := h.moneyService.GetProjectBySlug(c.Request.Context(), c.Param("project_id"), appmw.CurrentUser(c))
	respondMoney(c, resp, err)
}

func (h *Handler) GetMoneySnapshot(c *gin.Context) {
	resp, err := h.moneyService.Snapshot(c.Request.Context(), c.Param("project_id"))
	respondMoney(c, resp, err)
}

func (h *Handler) GetMoneyCragSnapshot(c *gin.Context) {
	resp, err := h.moneyService.CragSnapshot(c.Request.Context(), c.Param("project_id"))
	respondMoney(c, resp, err)
}

func (h *Handler) ListMoneyTrash(c *gin.Context) {
	resp, err := h.moneyService.ListTrash(c.Request.Context(), c.Param("project_id"))
	respondMoney(c, resp, err)
}

func (h *Handler) ListMoneyFeatures(c *gin.Context) {
	filter := models.MoneyFeatureFilter{FeatureType: c.Query("type"), Status: c.Query("status")}
	if bbox := c.Query("bbox"); bbox != "" {
		b, err := service.ParseBBox(bbox)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bbox"})
			return
		}
		filter.BBox = b
	}
	if updated := c.Query("updated_after"); updated != "" {
		t, err := time.Parse(time.RFC3339, updated)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid updated_after"})
			return
		}
		filter.UpdatedAfter = &t
	}
	resp, err := h.moneyService.ListFeatures(c.Request.Context(), c.Param("project_id"), filter)
	respondMoney(c, gin.H{"features": resp}, err)
}

func (h *Handler) CreateMoneyFeature(c *gin.Context) {
	var req models.MoneyFeatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	resp, err := h.moneyService.CreateFeature(c.Request.Context(), c.Param("project_id"), req, appmw.CurrentUser(c))
	if err != nil {
		respondMoney(c, nil, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) CreateMoneyArea(c *gin.Context) {
	var req models.MoneyCragAreaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	resp, err := h.moneyService.CreateArea(c.Request.Context(), c.Param("project_id"), req, appmw.CurrentUser(c))
	if err != nil {
		respondMoney(c, nil, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) CreateMoneyBoulder(c *gin.Context) {
	var req models.MoneyCragBoulderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	resp, err := h.moneyService.CreateBoulder(c.Request.Context(), c.Param("project_id"), req, appmw.CurrentUser(c))
	if err != nil {
		respondMoney(c, nil, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) CreateMoneyProblem(c *gin.Context) {
	var req models.MoneyCragProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	resp, err := h.moneyService.CreateProblem(c.Request.Context(), c.Param("project_id"), req, appmw.CurrentUser(c))
	if err != nil {
		respondMoney(c, nil, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) GetMoneyFeature(c *gin.Context) {
	resp, err := h.moneyService.GetFeatureDetail(c.Request.Context(), c.Param("feature_id"))
	respondMoney(c, resp, err)
}

func (h *Handler) UpdateMoneyFeature(c *gin.Context) {
	var req models.MoneyFeatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	resp, err := h.moneyService.UpdateFeature(c.Request.Context(), c.Param("feature_id"), req, appmw.CurrentUser(c))
	respondMoney(c, resp, err)
}

func (h *Handler) UpdateMoneyBoulderStatus(c *gin.Context) {
	var req models.MoneyBoulderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	resp, err := h.moneyService.UpdateBoulderStatus(c.Request.Context(), c.Param("feature_id"), req, appmw.CurrentUser(c))
	respondMoney(c, resp, err)
}

func (h *Handler) UpdateMoneyAreaGeometry(c *gin.Context) {
	var req models.MoneyAreaGeometryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	resp, err := h.moneyService.UpdateAreaGeometry(c.Request.Context(), c.Param("feature_id"), req, appmw.CurrentUser(c))
	respondMoney(c, resp, err)
}

func (h *Handler) ArchiveMoneyFeature(c *gin.Context) {
	mode := models.MoneyArchiveMode(strings.TrimSpace(c.Query("mode")))
	if c.Request.Body != nil && c.Request.ContentLength != 0 {
		var req models.MoneyArchiveFeatureRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		if req.Mode != "" {
			mode = req.Mode
		}
	}
	err := h.moneyService.ArchiveFeatureWithMode(c.Request.Context(), c.Param("feature_id"), mode, appmw.CurrentUser(c))
	respondMoney(c, gin.H{"status": "archived", "mode": mode}, err)
}

func (h *Handler) MoveMoneyFeatureParent(c *gin.Context) {
	var req models.MoneyMoveFeatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	resp, err := h.moneyService.MoveFeatureParent(c.Request.Context(), c.Param("feature_id"), req, appmw.CurrentUser(c))
	respondMoney(c, resp, err)
}

func (h *Handler) RestoreMoneyFeature(c *gin.Context) {
	err := h.moneyService.RestoreFeature(c.Request.Context(), c.Param("feature_id"), appmw.CurrentUser(c))
	respondMoney(c, gin.H{"status": "restored"}, err)
}

func (h *Handler) ListMoneyNotes(c *gin.Context) {
	resp, err := h.moneyService.ListNotes(c.Request.Context(), c.Param("feature_id"))
	respondMoney(c, gin.H{"notes": resp}, err)
}

func (h *Handler) ListMoneyProjectNotes(c *gin.Context) {
	resp, err := h.moneyService.ListProjectNotes(c.Request.Context(), c.Param("project_id"))
	respondMoney(c, gin.H{"notes": resp}, err)
}

func (h *Handler) CreateMoneyProjectNote(c *gin.Context) {
	var req models.MoneyNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	resp, err := h.moneyService.CreateProjectNote(c.Request.Context(), c.Param("project_id"), req, appmw.CurrentUser(c))
	if err != nil {
		respondMoney(c, nil, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) CreateMoneyNote(c *gin.Context) {
	var req models.MoneyNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	resp, err := h.moneyService.CreateNote(c.Request.Context(), c.Param("feature_id"), req, appmw.CurrentUser(c))
	if err != nil {
		respondMoney(c, nil, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) UpdateMoneyNote(c *gin.Context) {
	var req models.MoneyNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	resp, err := h.moneyService.UpdateNote(c.Request.Context(), c.Param("note_id"), req, appmw.CurrentUser(c))
	respondMoney(c, resp, err)
}

func (h *Handler) DeleteMoneyNote(c *gin.Context) {
	err := h.moneyService.DeleteNote(c.Request.Context(), c.Param("note_id"), appmw.CurrentUser(c))
	respondMoney(c, gin.H{"status": "deleted"}, err)
}

func (h *Handler) UploadMoneyImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	featureID := optionalForm(c, "feature_id")
	noteID := optionalForm(c, "note_id")
	blockKind := strings.TrimSpace(c.PostForm("block_kind"))
	resp, err := h.moneyService.StoreUploadWithKind(c.Request.Context(), c.Param("project_id"), featureID, noteID, blockKind, []byte(c.PostForm("metadata")), file, appmw.CurrentUser(c))
	if err != nil {
		respondMoney(c, nil, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) StreamMoneyUpload(c *gin.Context) {
	u, r, err := h.moneyService.OpenUpload(c.Request.Context(), c.Param("upload_id"))
	if err != nil {
		respondMoney(c, nil, err)
		return
	}
	defer r.Close()
	c.Header("Content-Disposition", "inline; filename=\""+strings.ReplaceAll(u.OriginalFilename, "\"", "")+"\"")
	c.DataFromReader(http.StatusOK, u.ByteSize, u.ContentType, r, nil)
}

func (h *Handler) GetMoneyUploadDownloadURL(c *gin.Context) {
	resp, err := h.moneyService.SignedUploadDownloadURL(c.Request.Context(), c.Param("upload_id"))
	respondMoney(c, resp, err)
}

func (h *Handler) UpdateMoneyUploadMetadata(c *gin.Context) {
	var req models.MoneyUploadMetadataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	resp, err := h.moneyService.UpdateUploadMetadata(c.Request.Context(), c.Param("upload_id"), req, appmw.CurrentUser(c))
	respondMoney(c, resp, err)
}

func (h *Handler) DeleteMoneyUpload(c *gin.Context) {
	err := h.moneyService.DeleteUpload(c.Request.Context(), c.Param("upload_id"), appmw.CurrentUser(c))
	respondMoney(c, gin.H{"status": "deleted"}, err)
}

func optionalForm(c *gin.Context, key string) *string {
	v := strings.TrimSpace(c.PostForm(key))
	if v == "" {
		return nil
	}
	return &v
}

func respondMoney(c *gin.Context, body interface{}, err error) {
	if err == nil {
		c.JSON(http.StatusOK, body)
		return
	}
	status := http.StatusInternalServerError
	msg := "Request failed"
	if errors.Is(err, service.ErrMoneyInvalidInput) {
		status = http.StatusBadRequest
		msg = "Invalid request"
	}
	if errors.Is(err, service.ErrMoneyForbidden) {
		status = http.StatusForbidden
		msg = "Forbidden"
	}
	if dberrors.IsNotFound(err) {
		status = http.StatusNotFound
		msg = "Not found"
	}
	c.JSON(status, gin.H{"error": msg})
}
