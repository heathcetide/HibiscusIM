package handlers

import (
	"HibiscusIM/internal/models"
	"HibiscusIM/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) handleWriteQuestionnaire(context *gin.Context) {
	var req models.QuestionnaireSubmitRequest
	if err := context.ShouldBindJSON(&req); err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user := models.CurrentUser(context)
	res, err := models.SubmitUserResponse(h.db, user.ID, req.QuestionnaireID, req.Answers)
	if err != nil {
		response.Fail(context, "error", gin.H{"error": err.Error()})
		return
	}
	if res != nil {
		response.Success(context, "success", gin.H{"data": res})
	}
	response.Fail(context, "failed", gin.H{})
}

func (h *Handlers) handleGetQuestionResponseById(context *gin.Context) {
	user := models.CurrentUser(context)
	questionnaireID := context.DefaultQuery("questionnaireId", "")
	if questionnaireID == "" {
		response.Fail(context, "error", gin.H{"error": "questionnaireId is empty"})
		return
	}
	questionnaireIDInt, err := strconv.ParseUint(questionnaireID, 10, 32)
	if err != nil {
		response.Fail(context, "error", "Invalid questionnaire ID")
		return
	}
	responses, err := models.GetResponsesByQuestionnaire(h.db, uint(user.ID), uint(questionnaireIDInt))
	if err != nil {
		response.Fail(context, "error", gin.H{"error": err.Error()})
		return
	}
	response.Success(context, "success", gin.H{"responses": responses})
}
