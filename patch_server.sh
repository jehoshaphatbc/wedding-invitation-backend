#!/bin/bash
sed -i '' '/admin.POST("\/users", userHandler.CreateUser)/i \
		admin.GET("/users/trash", userHandler.GetTrashedUsers)\
' pkg/server/server.go

sed -i '' '/admin.DELETE("\/users\/:id", userHandler.DeleteUser)/a \
		admin.POST("/users/:id/restore", userHandler.RestoreUser)\
		admin.DELETE("/users/:id/force", userHandler.ForceDeleteUser)\
' pkg/server/server.go

sed -i '' '/admin.POST("\/roles", roleHandler.CreateRole)/i \
		admin.GET("/roles/trash", roleHandler.GetTrashedRoles)\
' pkg/server/server.go

sed -i '' '/admin.DELETE("\/roles\/:id", roleHandler.DeleteRole)/a \
		admin.POST("/roles/:id/restore", roleHandler.RestoreRole)\
		admin.DELETE("/roles/:id/force", roleHandler.ForceDeleteRole)\
' pkg/server/server.go

