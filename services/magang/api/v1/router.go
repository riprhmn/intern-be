package v1

import (
	"magang-be/services/magang/internal/database"
	"magang-be/services/magang/internal/handler"
	"magang-be/services/magang/internal/service"
	"magang-be/services/magang/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine, itemSvc *service.ItemService, userSvc *service.UserService, changeNoteSvc *service.ChangeNoteService, authHandler *handler.AuthHandler, jwtSecret string) {
	itemHandler := handler.NewItemHandler(itemSvc)
	userHandler := handler.NewUserHandler(userSvc)
	approvalSettingHandler := handler.NewApprovalSettingHandler(service.NewApprovalSettingService(database.DB))
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
				users.GET("/:id", userHandler.GetByID)
				users.POST("", userHandler.Create)
				users.PUT("/:id", userHandler.Update)
				users.DELETE("/:id", userHandler.Delete)
			}

			approvalSettings := protected.Group("/approval-settings", middleware.RequireRole("admin"))
			{
				approvalSettings.GET("/options", approvalSettingHandler.Options)
				approvalSettings.GET("/mappings", approvalSettingHandler.GetMapping)
				approvalSettings.PUT("/mappings/:approvalType/:sectionCode", approvalSettingHandler.ReplaceMapping)
				approvalSettings.DELETE("/mappings/:approvalType/:sectionCode", approvalSettingHandler.DeleteMapping)
				approvalSettings.GET("/delegations", approvalSettingHandler.GetDelegations)
				approvalSettings.POST("/delegations", approvalSettingHandler.CreateDelegation)
				approvalSettings.PUT("/delegations/:id", approvalSettingHandler.UpdateDelegation)
				approvalSettings.DELETE("/delegations/:id", approvalSettingHandler.DeactivateDelegation)
			}

			// Change Notes routes
			changeNotes := protected.Group("/change-notes")
			{
				changeNotes.GET("", changeNoteHandler.GetAll)
				changeNotes.GET("/:id", changeNoteHandler.GetByID)
				changeNotes.POST("", changeNoteHandler.Create)
			}
		}
	}
}
