package handlers

import (
	hibiscusIM "HibiscusIM"
	"HibiscusIM/internal/apidocs"
	"HibiscusIM/internal/models"
	"HibiscusIM/pkg/config"
	"HibiscusIM/pkg/middleware"
	"HibiscusIM/pkg/notification"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handlers struct {
	db *gorm.DB
}

func NewHandlers(db *gorm.DB) *Handlers {
	return &Handlers{
		db: db,
	}
}

func (h *Handlers) Register(engine *gin.Engine) {
	r := engine.Group(config.GlobalConfig.APIPrefix)

	// Register Global Singleton DB
	r.Use(middleware.InjectDB(h.db))
	// Register System Module Routes
	h.registerSystemRoutes(r)

	// Register Business Module Routes
	h.registerAuthRoutes(r)
	h.registerNotificationRoutes(r)
	h.registerGroupRoutes(r)

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

func (h *Handlers) GetObjs() []hibiscusIM.WebObject {
	return []hibiscusIM.WebObject{
		{
			Group:       "hibiscusIM",
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
	admins := []models.AdminObject{
		{
			Model:       &notification.InternalNotification{},
			Group:       "Business",
			Name:        "InternalNotification",
			Desc:        "This is a notification used to notify the user of the system.",
			Shows:       []string{"ID", "Title", "Read", "CreatedAt"},
			Editables:   []string{"ID", "UserID", "Title", "Content", "Read", "CreatedAt"},
			Orderables:  []string{"CreatedAt"},
			Searchables: []string{"Title"},
			Icon:        &models.AdminIcon{SVG: string(iconInternalNotification)},
		},
	}
	models.RegisterAdmins(router, h.db, append(adminObjs, admins...))
}
