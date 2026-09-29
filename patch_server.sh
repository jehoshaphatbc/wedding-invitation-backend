#!/bin/bash
sed -i '' 's/admin.GET("\/users\/trash", userHandler.GetTrashedUsers)/admin.GET("\/users\/trash", middleware.SuperAdminOnly(), userHandler.GetTrashedUsers)/' pkg/server/server.go
sed -i '' 's/admin.POST("\/users\/:id\/restore", userHandler.RestoreUser)/admin.POST("\/users\/:id\/restore", middleware.SuperAdminOnly(), userHandler.RestoreUser)/' pkg/server/server.go
sed -i '' 's/admin.DELETE("\/users\/:id\/force", userHandler.ForceDeleteUser)/admin.DELETE("\/users\/:id\/force", middleware.SuperAdminOnly(), userHandler.ForceDeleteUser)/' pkg/server/server.go

sed -i '' 's/admin.GET("\/roles\/trash", roleHandler.GetTrashedRoles)/admin.GET("\/roles\/trash", middleware.SuperAdminOnly(), roleHandler.GetTrashedRoles)/' pkg/server/server.go
sed -i '' 's/admin.POST("\/roles\/:id\/restore", roleHandler.RestoreRole)/admin.POST("\/roles\/:id\/restore", middleware.SuperAdminOnly(), roleHandler.RestoreRole)/' pkg/server/server.go
sed -i '' 's/admin.DELETE("\/roles\/:id\/force", roleHandler.ForceDeleteRole)/admin.DELETE("\/roles\/:id\/force", middleware.SuperAdminOnly(), roleHandler.ForceDeleteRole)/' pkg/server/server.go
