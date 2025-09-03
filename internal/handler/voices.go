package handlers

import (
	"HibiscusIM/internal/models"
	"HibiscusIM/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 获取所有录音提示（每个待录音的句子）
func (h *Handlers) handleGetRecordingPrompts(c *gin.Context) {
	var prompts []models.RecordingPrompt
	if err := h.db.Find(&prompts).Error; err != nil {
		response.Fail(c, "can not find recording prompt records", nil)
		return
	}
	response.Success(c, "get recording prompts", prompts)
}

// 确认上传并保存录音信息
func (h *Handlers) ConfirmRecordingUpload(c *gin.Context) {
	var req struct {
		PromptID   uint   `json:"promptId"`
		FileUrl    string `json:"fileUrl"`
		Format     string `json:"format"`
		DurationMs int    `json:"durationMs"`
		Checksum   string `json:"checksum"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前用户
	user := models.CurrentUser(c)

	recording := models.Recording{
		UserID:     user.ID,
		PromptID:   req.PromptID,
		FileURL:    req.FileUrl,
		Format:     req.Format,
		DurationMs: req.DurationMs,
		Checksum:   req.Checksum,
		Status:     "uploaded",
	}

	if err := h.db.Create(&recording).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"recordingId": recording.ID})
}

// 获取生成的音频或合成结果
func (h *Handlers) GetVoiceJobResult(c *gin.Context) {
	jobID := c.Param("jobId")

	var job models.VoiceJob
	if err := h.db.First(&job, jobID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	if job.ResultURL == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Result not ready"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"resultUrl": job.ResultURL})
}
