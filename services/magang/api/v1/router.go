package v1

import (
	"magang-be/services/magang/internal/handler"
	"magang-be/services/magang/internal/service"
	"magang-be/services/magang/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine, itemSvc *service.ItemService, userSvc *service.UserService, changeNoteSvc *service.ChangeNoteService, ancrHandler *handler.ANCRHandler, approvalSettingHandler *handler.ApprovalSettingHandler, authHandler *handler.AuthHandler, jwtSecret string) {
	itemHandler := handler.NewItemHandler(itemSvc)
	userHandler := handler.NewUserHandler(userSvc)
	changeNoteHandler := handler.NewChangeNoteHandler(changeNoteSvc, userSvc)

	v1 := r.Group("/api/v1/magang")
	{
		protected := v1.Group("", middleware.JWTAuth(jwtSecret))
		{
			// Auth routes
			authGroup := protected.Group("/auth")
			{
				authGroup.GET("/me", authHandler.GetMe)
			}

			// Items routes
			items := protected.Group("/items")
			{
				items.GET("", itemHandler.GetAll)
				items.GET("/:id", itemHandler.GetByID)
				items.POST("", itemHandler.Create)
				items.PUT("/:id", itemHandler.Update)
				items.DELETE("/:id", itemHandler.Delete)
			}

			// Users routes
			users := protected.Group("/users")
			{
				users.GET("", userHandler.GetAll)
				users.GET("/organization-options", userHandler.OrganizationOptions)
				users.POST("/organization-options/validate", userHandler.ValidateOrganizationSelection)
				users.GET("/:id", userHandler.GetByID)
				users.POST("", middleware.RequireRole("admin"), userHandler.Create)
				users.PUT("/:id", middleware.RequireRole("admin"), userHandler.Update)
				users.DELETE("/:id", middleware.RequireRole("admin"), userHandler.Delete)
			}

			approvalSettings := protected.Group("/approval-settings")
			{
				approvalSettings.GET("/options", approvalSettingHandler.Options)
				approvalSettings.GET("/mappings", approvalSettingHandler.GetMapping)
				approvalSettings.GET("/delegations", approvalSettingHandler.GetDelegations)

				approvalSettings.PUT("/mappings/:approvalType/:sectionCode", middleware.RequireRole("admin"), approvalSettingHandler.ReplaceMapping)
				approvalSettings.DELETE("/mappings/:approvalType/:sectionCode", middleware.RequireRole("admin"), approvalSettingHandler.DeleteMapping)
				approvalSettings.POST("/delegations", middleware.RequireRole("admin"), approvalSettingHandler.CreateDelegation)
				approvalSettings.PUT("/delegations/:id", middleware.RequireRole("admin"), approvalSettingHandler.UpdateDelegation)
				approvalSettings.DELETE("/delegations/:id", middleware.RequireRole("admin"), approvalSettingHandler.DeactivateDelegation)
			}

			templates := protected.Group("/document-templates")
			templates.GET("", changeNoteHandler.Templates)
			templates.GET("/:templateID", changeNoteHandler.Templates)
			templates.PUT("", changeNoteHandler.Templates)
			templates.PUT("/:templateID", changeNoteHandler.Templates)
			templates.DELETE("/:templateID", changeNoteHandler.Templates)

			// Change Notes routes
			changeNotes := protected.Group("/change-notes")
			{
				changeNotes.GET("/summary", changeNoteHandler.Summary)
				changeNotes.GET("/approvers", changeNoteHandler.Approvers)
				changeNotes.GET("", changeNoteHandler.GetAll)
				changeNotes.GET("/:id", changeNoteHandler.GetByID)
				changeNotes.POST("", changeNoteHandler.Create)
				changeNotes.PUT("/:id", changeNoteHandler.Update)
			}

			// Dedicated ANCR routes (separate from Change Note)
			ancr := protected.Group("/ancr")
			{
				ancr.GET("", ancrHandler.GetAll)
				ancr.GET("/:id", ancrHandler.GetByID)
				ancr.POST("", ancrHandler.Create)
				ancr.PUT("/:id", ancrHandler.Update)
				ancr.DELETE("/:id", ancrHandler.Delete)
			}
		}
	}
}
