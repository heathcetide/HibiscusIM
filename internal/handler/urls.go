package handlers

import (
	hibiscusIM "HibiscusIM"
	"HibiscusIM/internal/apidocs"
	"HibiscusIM/internal/models"
	"HibiscusIM/pkg/config"
	"HibiscusIM/pkg/logger"
	"HibiscusIM/pkg/middleware"
	"HibiscusIM/pkg/notification"
	"HibiscusIM/pkg/search"
	"HibiscusIM/pkg/websocket"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handlers struct {
	db            *gorm.DB
	wsHub         *websocket.Hub
	searchHandler *search.SearchHandlers
}

func NewHandlers(db *gorm.DB) *Handlers {
	wsConfig := websocket.LoadConfigFromEnv()
	wsHub := websocket.NewHub(wsConfig)
	var searchHandler *search.SearchHandlers
	if config.GlobalConfig.SearchEnabled {
		engine, err := search.New(
			search.Config{
				IndexPath:    config.GlobalConfig.SearchPath,
				QueryTimeout: 5 * time.Second,
				BatchSize:    config.GlobalConfig.SearchBatchSize,
			},
			search.BuildIndexMapping(""),
		)
		if err != nil {
			log.Fatalf("Failed to initialize search engine: %v", err)
		}
		searchHandler = search.NewSearchHandlers(engine)
	}

	return &Handlers{
		db:            db,
		wsHub:         wsHub,
		searchHandler: searchHandler,
	}
}

func (h *Handlers) Register(engine *gin.Engine) {
	r := engine.Group(config.GlobalConfig.APIPrefix)

	// Register Global Singleton DB
	r.Use(middleware.InjectDB(h.db))
	if config.GlobalConfig.SearchEnabled {
		h.searchHandler.RegisterSearchRoutes(r)
	} else {
		logger.Info("Search API is disabled")
	}
	// Register System Module Routes
	h.registerSystemRoutes(r)

	// Register Business Module Routes
	h.registerAuthRoutes(r)
	h.registerNotificationRoutes(r)
	h.registerGroupRoutes(r)
	h.registerWebSocketRoutes(r)
	h.registerVoicesRoutes(r)
	h.registerQuestionRoutes(r)

	objs := h.GetObjs()
	hibiscusIM.RegisterObjects(r, objs)
	if config.GlobalConfig.DocsPrefix != "" {
		var objDocs []apidocs.WebObjectDoc
		for _, obj := range objs {
			objDocs = append(objDocs, apidocs.GetWebObjectDocDefine(config.GlobalConfig.APIPrefix, obj))
		}
		apidocs.RegisterHandler(config.GlobalConfig.DocsPrefix, engine, h.GetDocs(), objDocs, h.db)
	}
	if config.GlobalConfig.AdminPrefix != "" {
		admin := r.Group(config.GlobalConfig.AdminPrefix)
		h.RegisterAdmin(admin)
	}
}

// User Module
func (h *Handlers) registerAuthRoutes(r *gin.RouterGroup) {
	auth := r.Group(config.GlobalConfig.AuthPrefix)
	{
		// register
		auth.GET("/register", h.handleUserSignupPage)

		auth.POST("/register", h.handleUserSignup)

		auth.POST("/register/email", h.handleUserSignupByEmail)

		auth.POST("/send/email", h.handleSendEmailCode)

		// login
		auth.GET("/login", h.handleUserSigninPage)

		auth.POST("/login", h.handleUserSignin)

		auth.POST("/login/email", h.handleUserSigninByEmail)

		// logout
		auth.GET("/logout", models.AuthRequired, h.handleUserLogout)

		auth.GET("/info", models.AuthRequired, h.handleUserInfo)

		auth.GET("/reset-password", h.handleUserResetPasswordPage)

		// update
		auth.PUT("/update", models.AuthRequired, h.handleUserUpdate)

		auth.PUT("/update/preferences", models.AuthRequired, h.handleUserUpdatePreferences)

		auth.POST("/update/basic/info", models.AuthRequired, h.handleUserUpdateBasicInfo)
	}
}

func (h *Handlers) registerNotificationRoutes(r *gin.RouterGroup) {
	notificationGroup := r.Group("notification")
	{
		notificationGroup.GET("unread-count", models.AuthRequired, h.handleUnReadNotificationCount)

		notificationGroup.GET("", models.AuthRequired, h.handleListNotifications)

		notificationGroup.POST("readAll", models.AuthRequired, h.handleAllNotifications)

		notificationGroup.PUT("/read/:id", models.AuthRequired, h.handleMarkNotificationAsRead)

		notificationGroup.DELETE("/:id", models.AuthRequired, h.handleDeleteNotification)
	}
}

func (h *Handlers) registerSystemRoutes(r *gin.RouterGroup) {
	system := r.Group("system")
	{
		system.POST("/rate-limiter/config", h.UpdateRateLimiterConfig)

		system.GET("/health", h.HealthCheck)
	}
}

func (h *Handlers) registerGroupRoutes(r *gin.RouterGroup) {
	group := r.Group("group")
	group.OPTIONS("/*cors", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.AbortWithStatus(204)
	})
	group.Use(models.AuthRequired)
	{
		group.POST("/", h.CreateGroup)

		group.GET("/", h.ListGroups)

		group.GET("/:id", h.GetGroup)

		group.PUT("/:id", h.UpdateGroup)

		group.DELETE("/:id", h.DeleteGroup)
	}
}

func (h *Handlers) registerQuestionRoutes(r *gin.RouterGroup) {
	question := r.Group("question")
	question.Use(models.AuthRequired)
	{
		question.POST("/", h.handleWriteQuestionnaire)

		question.GET("/responses", h.handleGetQuestionResponseById)
	}
}

func (h *Handlers) registerVoicesRoutes(r *gin.RouterGroup) {
	voices := r.Group("voices")
	{
		voices.GET("/", h.handleGetRecordingPrompts)
	}
}

func (h *Handlers) GetObjs() []hibiscusIM.WebObject {
	return []hibiscusIM.WebObject{
		{
			Group:       "hibiscus",
			Desc:        "用户",
			Model:       models.User{},
			Name:        "user",
			Filterables: []string{"UpdateAt", "CreatedAt"},
			Editables:   []string{"Email", "Phone", "FirstName", "LastName", "DisplayName", "IsSuperUser", "Enabled"},
			Searchables: []string{},
			Orderables:  []string{"UpdatedAt"},
			GetDB: func(c *gin.Context, isCreate bool) *gorm.DB {
				if isCreate {
					return h.db
				}
				return h.db.Where("deleted_at", nil)
			},
			BeforeCreate: func(db *gorm.DB, ctx *gin.Context, vptr any) error {
				return nil
			},
		},
	}
}

func (h *Handlers) RegisterAdmin(router *gin.RouterGroup) {
	adminObjs := models.GetHibiscusAdminObjects()
	iconInternalNotification, _ := hibiscusIM.EmbedStaticAssets.ReadFile("static/img/icon_internal_notification.svg")
	iconOperatorLog, _ := hibiscusIM.EmbedStaticAssets.ReadFile("static/img/icon_operator_log.svg")
	iconQuestion, _ := hibiscusIM.EmbedStaticAssets.ReadFile("static/img/icon_question.svg")
	iconQuestionnaire, _ := hibiscusIM.EmbedStaticAssets.ReadFile("static/img/icon_questionnaire.svg")
	iconAnswer, _ := hibiscusIM.EmbedStaticAssets.ReadFile("static/img/icon_answer.svg")
	iconQuestionnaireResponse, _ := hibiscusIM.EmbedStaticAssets.ReadFile("static/img/icon_questionnaire_response.svg")
	iconRecording, _ := hibiscusIM.EmbedStaticAssets.ReadFile("static/img/icon_recording.svg")
	iconRecordingPrompt, _ := hibiscusIM.EmbedStaticAssets.ReadFile("static/img/icon_recording_prompt.svg")
	iconVoiceJob, _ := hibiscusIM.EmbedStaticAssets.ReadFile("static/img/icon_voice_job.svg")
	admins := []models.AdminObject{
		{
			Model:       &notification.InternalNotification{},
			Group:       "System",
			Name:        "InternalNotification",
			Desc:        "This is a notification used to notify the user of the system.",
			Shows:       []string{"ID", "Title", "Read", "CreatedAt"},
			Editables:   []string{"ID", "UserID", "Title", "Content", "Read", "CreatedAt"},
			Orderables:  []string{"CreatedAt"},
			Searchables: []string{"Title"},
			Icon:        &models.AdminIcon{SVG: string(iconInternalNotification)},
		},
		{
			Model:       &middleware.OperationLog{},                                  // 关联模型 OperationLog
			Group:       "System",                                                    // 业务组
			Name:        "Operation Log",                                             // 管理员后台展示的名称
			Desc:        "Logs the operations performed by users in the system.",     // 描述
			Shows:       []string{"ID", "Username", "Action", "Target", "CreatedAt"}, // 显示的字段
			Editables:   []string{"Action", "Target", "Details"},                     // 可编辑字段
			Orderables:  []string{"CreatedAt"},                                       // 可排序字段
			Searchables: []string{"Username", "Action", "Target"},                    // 可搜索字段
			Icon:        &models.AdminIcon{SVG: string(iconOperatorLog)},             // 图标
		},
		{
			Model:       &models.Question{},                           // 关联 Question 模型
			Group:       "Survey",                                     // 业务组
			Name:        "Question",                                   // 管理员后台展示的名称
			Desc:        "This is the question in a questionnaire.",   // 描述
			Shows:       []string{"ID", "Text", "Type", "Options"},    // 显示的字段
			Editables:   []string{"Text", "Type", "Options"},          // 可编辑字段
			Orderables:  []string{"CreatedAt"},                        // 可排序字段
			Searchables: []string{"Text", "Type"},                     // 可搜索字段
			Icon:        &models.AdminIcon{SVG: string(iconQuestion)}, // 图标
		},
		{
			Model:       &models.Questionnaire{},                               // 关联 Questionnaire 模型
			Group:       "Survey",                                              // 业务组
			Name:        "Questionnaire",                                       // 管理员后台展示的名称
			Desc:        "This is a questionnaire, a collection of questions.", // 描述
			Shows:       []string{"ID", "Title", "Description", "CreatedAt"},   // 显示的字段
			Editables:   []string{"Title", "Description"},                      // 可编辑字段
			Orderables:  []string{"CreatedAt"},                                 // 可排序字段
			Searchables: []string{"Title", "Description"},                      // 可搜索字段
			Icon:        &models.AdminIcon{SVG: string(iconQuestionnaire)},     // 图标
		},
		{
			Model:       &models.Answer{},                                                         // 关联 Answer 模型
			Group:       "Survey",                                                                 // 业务组
			Name:        "Answer",                                                                 // 管理员后台展示的名称
			Desc:        "This is the answer provided by the user for a specific question.",       // 描述
			Shows:       []string{"ID", "ResponseID", "QuestionID", "AnswerText", "AnswerOption"}, // 显示的字段
			Editables:   []string{"ResponseID", "QuestionID", "AnswerText", "AnswerOption"},       // 可编辑字段
			Orderables:  []string{"CreatedAt"},                                                    // 可排序字段
			Searchables: []string{"AnswerText", "AnswerOption"},                                   // 可搜索字段
			Icon:        &models.AdminIcon{SVG: string(iconAnswer)},                               // 图标
		},
		{
			Model:       &models.QuestionnaireResponse{},                           // 关联 QuestionnaireResponse 模型
			Group:       "Survey",                                                  // 业务组
			Name:        "Questionnaire Response",                                  // 管理员后台展示的名称
			Desc:        "This records the responses of users to a questionnaire.", // 描述
			Shows:       []string{"ID", "UserID", "QuestionnaireID", "CreatedAt"},  // 显示的字段
			Editables:   []string{"UserID", "QuestionnaireID"},                     // 可编辑字段
			Orderables:  []string{"CreatedAt"},                                     // 可排序字段
			Searchables: []string{"UserID", "QuestionnaireID"},                     // 可搜索字段
			Icon:        &models.AdminIcon{SVG: string(iconQuestionnaireResponse)}, // 图标
		},
		{
			Model:       &models.RecordingPrompt{},                                            // 关联 RecordingPrompt 模型
			Group:       "Recording",                                                          // 业务组
			Name:        "Recording Prompt",                                                   // 管理员后台展示名称
			Desc:        "This is a recording prompt, a sentence to be recorded by the user.", // 描述
			Shows:       []string{"ID", "Text", "Order", "CreatedAt"},
			Editables:   []string{"Text", "Order"},
			Orderables:  []string{"CreatedAt"},
			Searchables: []string{"Text"},
			Icon:        &models.AdminIcon{SVG: string(iconRecordingPrompt)}, // 图标
		},
		{
			Model:       &models.Recording{},                                    // 关联 Recording 模型
			Group:       "Recording",                                            // 业务组
			Name:        "Recording",                                            // 管理员后台展示名称
			Desc:        "This records the user’s voice for a specific prompt.", // 描述
			Shows:       []string{"ID", "UserID", "PromptID", "FileURL", "Format", "DurationMs", "Status", "CreatedAt"},
			Editables:   []string{"UserID", "PromptID", "FileURL", "Format", "DurationMs", "Status"},
			Orderables:  []string{"CreatedAt"},
			Searchables: []string{"FileURL", "Status"},
			Icon:        &models.AdminIcon{SVG: string(iconRecording)}, // 图标
		},
		{
			Model:       &models.VoiceJob{},                                            // 关联 VoiceJob 模型
			Group:       "Recording",                                                   // 业务组
			Name:        "Voice Job",                                                   // 管理员后台展示名称
			Desc:        "This represents a voice job for processing user recordings.", // 描述
			Shows:       []string{"ID", "UserID", "Status", "Progress", "CreatedAt"},
			Editables:   []string{"UserID", "Status", "Progress"},
			Orderables:  []string{"CreatedAt"},
			Searchables: []string{"Status", "Progress"},
			Icon:        &models.AdminIcon{SVG: string(iconVoiceJob)}, // 图标
		},
	}
	models.RegisterAdmins(router, h.db, append(adminObjs, admins...))
}

// registerWebSocketRoutes 注册WebSocket路由
func (h *Handlers) registerWebSocketRoutes(r *gin.RouterGroup) {
	wsHandler := websocket.NewHandler(h.wsHub)

	// WebSocket连接端点
	r.GET("/ws", models.AuthRequired, wsHandler.HandleWebSocket)

	// WebSocket管理API端点
	wsGroup := r.Group("/ws")
	wsGroup.Use(models.AuthRequired)
	{
		wsGroup.GET("/stats", wsHandler.GetStats)
		wsGroup.GET("/health", wsHandler.HealthCheck)
		wsGroup.GET("/user/:user_id", wsHandler.GetUserStats)
		wsGroup.GET("/group/:group", wsHandler.GetGroupStats)
		wsGroup.POST("/message", wsHandler.SendMessage)
		wsGroup.POST("/broadcast", wsHandler.BroadcastMessage)
		wsGroup.DELETE("/user/:user_id", wsHandler.DisconnectUser)
		wsGroup.DELETE("/group/:group", wsHandler.DisconnectGroup)
	}
}
