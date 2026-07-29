package routes

import (
	"strconv"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	ratelimitmiddleware "github.com/Wei-Shaw/sub2api/internal/middleware"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RegisterUserRoutes 注册用户相关路由（需要认证）
func RegisterUserRoutes(
	v1 *gin.RouterGroup,
	h *handler.Handlers,
	jwtAuth middleware.JWTAuthMiddleware,
	settingService *service.SettingService,
	redisClient *redis.Client,
) {
	authenticated := v1.Group("")
	authenticated.Use(gin.HandlerFunc(jwtAuth))
	authenticated.Use(middleware.BackendModeUserGuard(settingService))
	{
		// 用户接口
		user := authenticated.Group("/user")
		{
			user.GET("/profile", h.User.GetProfile)
			user.PUT("/password", h.User.ChangePassword)
			user.PUT("", h.User.UpdateProfile)
			user.GET("/aff", h.User.GetAffiliate)
			user.POST("/aff/transfer", h.User.TransferAffiliateQuota)
			user.POST("/account-bindings/email/send-code", h.User.SendEmailBindingCode)
			user.POST("/account-bindings/email", h.User.BindEmailIdentity)
			user.DELETE("/account-bindings/:provider", h.User.UnbindIdentity)
			user.POST("/auth-identities/bind/start", h.User.StartIdentityBinding)
			user.GET("/api-keys/:id/usage/daily", h.Usage.GetMyAPIKeyDailyUsage)
			user.GET("/platform-quotas", h.User.GetMyPlatformQuotas)

			// 通知邮箱管理
			notifyEmail := user.Group("/notify-email")
			{
				notifyEmail.POST("/send-code", h.User.SendNotifyEmailCode)
				notifyEmail.POST("/verify", h.User.VerifyNotifyEmail)
				notifyEmail.PUT("/toggle", h.User.ToggleNotifyEmail)
				notifyEmail.DELETE("", h.User.RemoveNotifyEmail)
			}

			// TOTP 双因素认证
			totp := user.Group("/totp")
			{
				totp.GET("/status", h.Totp.GetStatus)
				totp.GET("/verification-method", h.Totp.GetVerificationMethod)
				totp.POST("/send-code", h.Totp.SendVerifyCode)
				totp.POST("/setup", h.Totp.InitiateSetup)
				totp.POST("/enable", h.Totp.Enable)
				totp.POST("/disable", h.Totp.Disable)
			}
		}

		// API Key管理
		keys := authenticated.Group("/keys")
		{
			keys.GET("", h.APIKey.List)
			keys.GET("/:id", h.APIKey.GetByID)
			keys.POST("", h.APIKey.Create)
			keys.PUT("/:id", h.APIKey.Update)
			keys.DELETE("/:id", h.APIKey.Delete)
		}

		// 用户可用分组（非管理员接口）
		groups := authenticated.Group("/groups")
		{
			groups.GET("/available", h.APIKey.GetAvailableGroups)
			groups.GET("/rates", h.APIKey.GetUserGroupRates)
		}

		authenticated.GET("/model-market/catalog", h.ModelMarket.Catalog)

		workbench := authenticated.Group("/workbench")
		{
			rateLimiter := ratelimitmiddleware.NewRateLimiter(redisClient)
			failClose := ratelimitmiddleware.RateLimitOptions{FailureMode: ratelimitmiddleware.RateLimitFailClose}
			userFailClose := ratelimitmiddleware.RateLimitOptions{
				FailureMode: ratelimitmiddleware.RateLimitFailClose,
				KeyFunc: func(c *gin.Context) string {
					subject, ok := middleware.GetAuthSubjectFromContext(c)
					if !ok {
						return ""
					}
					return strconv.FormatInt(subject.UserID, 10)
				},
			}
			attachmentCreateIPLimit := rateLimiter.LimitWithOptions("workbench-attachment-create-ip", 30, time.Minute, failClose)
			attachmentCreateUserLimit := rateLimiter.LimitWithOptions("workbench-attachment-create-user", 12, time.Minute, userFailClose)
			attachmentUploadIPLimit := rateLimiter.LimitWithOptions("workbench-attachment-upload-ip", 30, time.Minute, failClose)
			attachmentUploadUserLimit := rateLimiter.LimitWithOptions("workbench-attachment-upload-user", 12, time.Minute, userFailClose)
			attachmentCompleteIPLimit := rateLimiter.LimitWithOptions("workbench-attachment-complete-ip", 30, time.Minute, failClose)
			attachmentCompleteUserLimit := rateLimiter.LimitWithOptions("workbench-attachment-complete-user", 12, time.Minute, userFailClose)
			attachmentDeleteIPLimit := rateLimiter.LimitWithOptions("workbench-attachment-delete-ip", 60, time.Minute, failClose)
			attachmentDeleteUserLimit := rateLimiter.LimitWithOptions("workbench-attachment-delete-user", 30, time.Minute, userFailClose)
			attachmentReadIPLimit := rateLimiter.LimitWithOptions("workbench-attachment-read-ip", 180, time.Minute, failClose)
			attachmentReadUserLimit := rateLimiter.LimitWithOptions("workbench-attachment-read-user", 120, time.Minute, userFailClose)

			workbench.GET("/models", h.Workbench.Models)
			workbench.POST("/models", h.Workbench.AddModel)
			workbench.DELETE("/models/:id", h.Workbench.HideModel)
			workbench.POST("/conversations", h.Workbench.CreateConversation)
			workbench.GET("/conversations", h.Workbench.ListConversations)
			workbench.GET("/conversations/:id", h.Workbench.GetConversation)
			workbench.PATCH("/conversations/:id", h.Workbench.UpdateConversation)
			workbench.DELETE("/conversations/:id", h.Workbench.DeleteConversation)
			workbench.POST("/conversations/:id/messages", middleware.RequestBodyLimit(512<<10), h.Workbench.CreateMessage)
			workbench.POST("/generations/:id/stream", h.Workbench.StreamGeneration)
			workbench.POST("/generations/:id/cancel", h.Workbench.CancelGeneration)
			workbench.POST("/attachments", attachmentCreateIPLimit, attachmentCreateUserLimit, middleware.RequestBodyLimit(32<<10), h.Workbench.CreateAttachmentUpload)
			workbench.PUT("/attachments/:id/content", attachmentUploadIPLimit, attachmentUploadUserLimit, middleware.RequestBodyLimit(service.WorkbenchMaxAttachmentBytes+1), h.Workbench.UploadAttachmentContent)
			workbench.POST("/attachments/:id/complete", attachmentCompleteIPLimit, attachmentCompleteUserLimit, h.Workbench.CompleteAttachmentUpload)
			workbench.GET("/attachments/:id/content", attachmentReadIPLimit, attachmentReadUserLimit, h.Workbench.AttachmentContent)
			workbench.DELETE("/attachments/:id", attachmentDeleteIPLimit, attachmentDeleteUserLimit, h.Workbench.DeleteAttachment)
		}

		// 用户可用渠道（非管理员接口）
		channels := authenticated.Group("/channels")
		{
			channels.GET("/available", h.AvailableChannel.List)
		}

		// 使用记录
		usage := authenticated.Group("/usage")
		{
			usage.GET("", h.Usage.List)
			usage.GET("/errors", h.Usage.ListErrors)
			usage.GET("/errors/:id", h.Usage.GetErrorDetail)
			usage.GET("/:id", h.Usage.GetByID)
			usage.GET("/stats", h.Usage.Stats)
			// User dashboard endpoints
			usage.GET("/dashboard/stats", h.Usage.DashboardStats)
			usage.GET("/dashboard/trend", h.Usage.DashboardTrend)
			usage.GET("/dashboard/models", h.Usage.DashboardModels)
			usage.GET("/dashboard/snapshot-v2", h.Usage.DashboardSnapshotV2)
			usage.POST("/dashboard/api-keys-usage", h.Usage.DashboardAPIKeysUsage)
		}

		// 公告（用户可见）
		announcements := authenticated.Group("/announcements")
		{
			announcements.GET("", h.Announcement.List)
			announcements.POST("/:id/read", h.Announcement.MarkRead)
		}

		// 卡密兑换
		redeem := authenticated.Group("/redeem")
		{
			redeem.POST("", h.Redeem.Redeem)
			redeem.GET("/history", h.Redeem.GetHistory)
		}

		imageGenerations := authenticated.Group("/image-generations")
		{
			imageGenerations.GET("/pricing", h.ImageGeneration.Pricing)
			imageGenerations.GET("/capabilities", h.ImageGeneration.Capabilities)
			imageGenerations.POST("", h.ImageGeneration.Create)
			imageGenerations.POST("/queued", h.ImageGeneration.CreateQueued)
			imageGenerations.GET("", h.ImageGeneration.List)
			imageGenerations.GET("/:id", h.ImageGeneration.Get)
			imageGenerations.GET("/:id/prompt-versions", h.ImageGeneration.PromptVersions)
			imageGenerations.DELETE("/:id", h.ImageGeneration.Delete)
			imageGenerations.GET("/:id/images/:index/preview", h.ImageGeneration.Preview)
			imageGenerations.GET("/:id/images/:index/download", h.ImageGeneration.Download)
		}

		// 用户订阅
		subscriptions := authenticated.Group("/subscriptions")
		{
			subscriptions.GET("", h.Subscription.List)
			subscriptions.GET("/active", h.Subscription.GetActive)
			subscriptions.GET("/progress", h.Subscription.GetProgress)
			subscriptions.GET("/summary", h.Subscription.GetSummary)
		}

		// 渠道监控（用户只读）
		monitors := authenticated.Group("/channel-monitors")
		{
			monitors.GET("", h.ChannelMonitor.List)
			monitors.GET("/:id/status", h.ChannelMonitor.GetStatus)
		}
	}
}
