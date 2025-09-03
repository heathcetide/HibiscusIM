package models

import (
	"time"

	"gorm.io/gorm"
)

type Question struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	QuestionnaireID uint      `json:"questionnaireId"`            // 问题所属的问卷ID
	Text            string    `json:"text" gorm:"size:512"`       // 问题文本
	Type            string    `json:"type" gorm:"size:50"`        // 问题类型（如：选择题、文本、打分题等）
	Options         []string  `json:"options,omitempty" gorm:"-"` // 可选项（如果是选择题）
	CreatedAt       time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

type Questionnaire struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" gorm:"size:255"`       // 问卷标题
	Description string    `json:"description" gorm:"size:255"` // 问卷描述
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

type QuestionnaireResponse struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	UserID          uint      `json:"userId"`          // 用户ID
	QuestionnaireID uint      `json:"questionnaireId"` // 问卷ID
	CreatedAt       time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

type Answer struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	ResponseID   uint   `json:"responseId"`                   // 问卷回答ID
	QuestionID   uint   `json:"questionId"`                   // 问题ID
	AnswerText   string `json:"answerText" gorm:"size:1024"`  // 用户的答案
	AnswerOption string `json:"answerOption" gorm:"size:256"` // 如果是选择题，存储用户选的选项
}

type QuestionnaireSubmitRequest struct {
	QuestionnaireID uint     `json:"questionnaireId"`
	Answers         []Answer `json:"answers"`
}

// CreateQuestionnaire 创建一个新的问卷
func CreateQuestionnaire(db *gorm.DB, title, description string) (*Questionnaire, error) {
	questionnaire := &Questionnaire{
		Title:       title,
		Description: description,
	}

	if err := db.Create(questionnaire).Error; err != nil {
		return nil, err
	}

	return questionnaire, nil
}

// GetQuestionnaire 获取单个问卷
func GetQuestionnaire(db *gorm.DB, id uint) (*Questionnaire, error) {
	var questionnaire Questionnaire
	if err := db.First(&questionnaire, id).Error; err != nil {
		return nil, err
	}
	return &questionnaire, nil
}

// GetAllQuestionnaires 获取所有问卷
func GetAllQuestionnaires(db *gorm.DB) ([]Questionnaire, error) {
	var questionnaires []Questionnaire
	if err := db.Find(&questionnaires).Error; err != nil {
		return nil, err
	}
	return questionnaires, nil
}

// UpdateQuestionnaire 更新问卷
func UpdateQuestionnaire(db *gorm.DB, id uint, title, description string) (*Questionnaire, error) {
	var questionnaire Questionnaire
	if err := db.First(&questionnaire, id).Error; err != nil {
		return nil, err
	}

	questionnaire.Title = title
	questionnaire.Description = description
	if err := db.Save(&questionnaire).Error; err != nil {
		return nil, err
	}

	return &questionnaire, nil
}

// DeleteQuestionnaire 删除问卷
func DeleteQuestionnaire(db *gorm.DB, id uint) error {
	if err := db.Delete(&Questionnaire{}, id).Error; err != nil {
		return err
	}
	return nil
}

// CreateQuestion 创建一个新的问题
func CreateQuestion(db *gorm.DB, questionnaireID uint, text, questionType string, options []string) (*Question, error) {
	question := &Question{
		QuestionnaireID: questionnaireID,
		Text:            text,
		Type:            questionType,
		Options:         options,
	}

	if err := db.Create(question).Error; err != nil {
		return nil, err
	}

	return question, nil
}

// GetQuestion 获取单个问题
func GetQuestion(db *gorm.DB, id uint) (*Question, error) {
	var question Question
	if err := db.First(&question, id).Error; err != nil {
		return nil, err
	}
	return &question, nil
}

// GetQuestionsByQuestionnaire 获取某个问卷下的所有问题
func GetQuestionsByQuestionnaire(db *gorm.DB, questionnaireID uint) ([]Question, error) {
	var questions []Question
	if err := db.Where("questionnaire_id = ?", questionnaireID).Find(&questions).Error; err != nil {
		return nil, err
	}
	return questions, nil
}

// UpdateQuestion 更新问题
func UpdateQuestion(db *gorm.DB, id uint, text, questionType string, options []string) (*Question, error) {
	var question Question
	if err := db.First(&question, id).Error; err != nil {
		return nil, err
	}

	question.Text = text
	question.Type = questionType
	question.Options = options
	if err := db.Save(&question).Error; err != nil {
		return nil, err
	}

	return &question, nil
}

// DeleteQuestion 删除问题
func DeleteQuestion(db *gorm.DB, id uint) error {
	if err := db.Delete(&Question{}, id).Error; err != nil {
		return err
	}
	return nil
}

// CreateAnswer 创建用户的答案
func CreateAnswer(db *gorm.DB, responseID, questionID uint, answerText, answerOption string) (*Answer, error) {
	answer := &Answer{
		ResponseID:   responseID,
		QuestionID:   questionID,
		AnswerText:   answerText,
		AnswerOption: answerOption,
	}

	if err := db.Create(answer).Error; err != nil {
		return nil, err
	}

	return answer, nil
}

// GetAnswersByQuestion 获取某个问题的所有答案
func GetAnswersByQuestion(db *gorm.DB, questionID uint) ([]Answer, error) {
	var answers []Answer
	if err := db.Where("question_id = ?", questionID).Find(&answers).Error; err != nil {
		return nil, err
	}
	return answers, nil
}

// GetAnswersByResponse 获取用户的所有答案
func GetAnswersByResponse(db *gorm.DB, responseID uint) ([]Answer, error) {
	var answers []Answer
	if err := db.Where("response_id = ?", responseID).Find(&answers).Error; err != nil {
		return nil, err
	}
	return answers, nil
}

// GetResponsesByUser 获取某个用户的所有问卷回答
func GetResponsesByUser(db *gorm.DB, userID uint) ([]QuestionnaireResponse, error) {
	var responses []QuestionnaireResponse
	if err := db.Where("user_id = ?", userID).Find(&responses).Error; err != nil {
		return nil, err
	}
	return responses, nil
}

// GetResponsesByQuestionnaire 获取某个问卷的所有回答
func GetResponsesByQuestionnaire(db *gorm.DB, userId, questionnaireID uint) ([]QuestionnaireResponse, error) {
	var responses []QuestionnaireResponse
	if err := db.Where("questionnaire_id = ? AND user_id = ?", questionnaireID, userId).Find(&responses).Error; err != nil {
		return nil, err
	}
	return responses, nil
}

// SubmitUserResponse 提交用户的问卷回答
func SubmitUserResponse(db *gorm.DB, userID, questionnaireID uint, answers []Answer) (*QuestionnaireResponse, error) {
	// 先创建问卷回答记录
	response := &QuestionnaireResponse{
		UserID:          userID,
		QuestionnaireID: questionnaireID,
	}
	// 保存问卷回答记录
	if err := db.Create(response).Error; err != nil {
		return nil, err
	}

	// 遍历用户的答案并保存到 Answer 表
	for _, answer := range answers {
		answer.ResponseID = response.ID // 关联到当前的问卷回答
		if err := db.Create(&answer).Error; err != nil {
			return nil, err
		}
	}
	// 返回提交的问卷回答记录
	return response, nil
}
