package listeners

import (
	"HibiscusIM/internal/models"
	"HibiscusIM/pkg/config"
	"HibiscusIM/pkg/logger"
	"HibiscusIM/pkg/notification"
	"HibiscusIM/pkg/util"

	"go.uber.org/zap"
)

func InitUserListeners() {
	// register initialized listener - Send Welcome Email And InternalNotification
	util.Sig().Connect(models.SigUserCreate, func(sender any, params ...any) {
		user := sender.(*models.User)
		if user.Email == "" {
			return
		}

		go func() {
			err := notification.NewMailNotification(config.GlobalConfig.Mail).SendWelcomeEmail(
				user.Email,
				user.DisplayName,
				"https://yourapp.com/verify?token=abc123",
			)
			if err != nil {
				logger.Warn("send mail failed", zap.Error(err))
			}
		}()
	})
}
