package seeds

import (
	"fmt"
	"log"

	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/models"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/auth"
)

var permissionList = []struct {
	Name        string
	DisplayName string
}{
	{"profile.view", "View Profile"},
	{"profile.update", "Update Profile"},
	{"user.view", "View Users"},
	{"user.create", "Create Users"},
	{"user.update", "Update Users"},
	{"user.delete", "Delete Users"},
	{"role.view", "View Roles"},
	{"role.create", "Create Roles"},
	{"role.update", "Update Roles"},
	{"role.delete", "Delete Roles"},
	{"role.assign", "Assign Roles"},
	{"permission.view", "View Permissions"},
	{"invitation.view", "View Invitations"},
	{"invitation.create", "Create Invitations"},
	{"invitation.update", "Update Invitations"},
	{"invitation.delete", "Delete Invitations"},
	{"invitation.publish", "Publish Invitations"},
	{"template.view", "View Templates"},
	{"template.create", "Create Templates"},
	{"template.update", "Update Templates"},
	{"template.delete", "Delete Templates"},
	{"rsvp.view", "View RSVPs"},
	{"rsvp.export", "Export RSVPs"},
	{"contact.view", "View Contacts"},
	{"contact.delete", "Delete Contacts"},
}

var roleSeeds = []struct {
	Name        string
	DisplayName string
	Description string
	IsSystem    bool
}{
	{"super_admin", "Super Admin", "Full system access", true},
	{"admin", "Admin", "Manage users, content and system", true},
	{"customer", "Customer", "Regular user with invitation access", true},
}

var adminPermissions = []string{
	"profile.view", "profile.update",
	"user.view", "user.create", "user.update",
	"role.view", "role.assign",
	"permission.view",
	"invitation.view", "invitation.create", "invitation.update", "invitation.delete", "invitation.publish",
	"template.view", "template.create", "template.update", "template.delete",
	"rsvp.view", "rsvp.export",
	"contact.view",
}

var customerPermissions = []string{
	"profile.view", "profile.update",
	"invitation.view", "invitation.create", "invitation.update", "invitation.delete", "invitation.publish",
	"template.view",
	"rsvp.view", "rsvp.export",
}

func Seed(db *gorm.DB, superAdminName, superAdminEmail, superAdminPassword string) {
	seedPermissions(db)
	seedRoles(db)
	seedRolePermissions(db)
	seedSuperAdmin(db, superAdminName, superAdminEmail, superAdminPassword)
}

func SeedSuperAdminOnly(db *gorm.DB, superAdminName, superAdminEmail, superAdminPassword string) {
	seedSuperAdmin(db, superAdminName, superAdminEmail, superAdminPassword)
}

func seedPermissions(db *gorm.DB) {
	for _, p := range permissionList {
		var existing models.Permission
		result := db.Where("name = ?", p.Name).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			db.Create(&models.Permission{
				Name:        p.Name,
				DisplayName: p.DisplayName,
			})
		}
	}
	fmt.Println("Permissions seeded successfully")
}

func seedRoles(db *gorm.DB) {
	for _, r := range roleSeeds {
		var existing models.Role
		result := db.Where("name = ?", r.Name).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			db.Create(&models.Role{
				Name:        r.Name,
				DisplayName: r.DisplayName,
				Description: &r.Description,
				IsSystem:    r.IsSystem,
			})
		}
	}
	fmt.Println("Roles seeded successfully")
}

func seedRolePermissions(db *gorm.DB) {
	rolePermissionMap := map[string][]string{
		"admin":    adminPermissions,
		"customer": customerPermissions,
	}

	for roleName, permNames := range rolePermissionMap {
		var role models.Role
		if err := db.Where("name = ?", roleName).First(&role).Error; err != nil {
			continue
		}

		db.Exec("DELETE FROM role_permissions WHERE role_id = ?", role.ID)

		for _, permName := range permNames {
			var perm models.Permission
			if err := db.Where("name = ?", permName).First(&perm).Error; err == nil {
				db.Exec("INSERT INTO role_permissions (role_id, permission_id, created_at) VALUES (?, ?, NOW())", role.ID, perm.ID)
			}
		}
	}

	var superAdminRole models.Role
	if err := db.Where("name = ?", "super_admin").First(&superAdminRole).Error; err == nil {
		db.Exec("DELETE FROM role_permissions WHERE role_id = ?", superAdminRole.ID)
		var allPerms []models.Permission
		db.Find(&allPerms)
		for _, p := range allPerms {
			db.Exec("INSERT INTO role_permissions (role_id, permission_id, created_at) VALUES (?, ?, NOW())", superAdminRole.ID, p.ID)
		}
	}

	fmt.Println("Role permissions seeded successfully")
}

func seedSuperAdmin(db *gorm.DB, name, email, password string) {
	var existing models.User
	result := db.Where("email = ?", email).First(&existing)
	if result.Error == nil {
		return
	}

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		log.Printf("Failed to hash super admin password: %v", err)
		return
	}

	user := &models.User{
		Name:         name,
		Email:        email,
		PasswordHash: hashedPassword,
		Status:       models.UserStatusActive,
	}
	db.Create(user)

	var superAdminRole models.Role
	if err := db.Where("name = ?", "super_admin").First(&superAdminRole).Error; err == nil {
		db.Exec("INSERT INTO user_roles (user_id, role_id, created_at) VALUES (?, ?, NOW())", user.ID, superAdminRole.ID)
	}

	db.Create(&models.ClientProfile{
		UserID: user.ID,
	})

	fmt.Printf("Super admin created: %s\n", email)
}
