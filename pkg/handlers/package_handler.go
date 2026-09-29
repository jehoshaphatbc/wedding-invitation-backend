package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/response"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/services"
)

type PackageHandler struct {
	packageService *services.PackageService
}

func NewPackageHandler(packageService *services.PackageService) *PackageHandler {
	return &PackageHandler{packageService: packageService}
}

func (h *PackageHandler) CreatePackage(c *gin.Context) {
	var req services.PackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body: " + err.Error())
		return
	}

	pkg, err := h.packageService.CreatePackage(req)
	if err != nil {
		response.InternalServerError(c, "Failed to create package: " + err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Package created successfully.", pkg)
}

func (h *PackageHandler) GetAllPackages(c *gin.Context) {
	packages, err := h.packageService.GetAllPackages()
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve packages: " + err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Packages retrieved successfully.", packages)
}

func (h *PackageHandler) GetPackage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid package ID.")
		return
	}

	pkg, err := h.packageService.GetPackageByID(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Package retrieved successfully.", pkg)
}

func (h *PackageHandler) UpdatePackage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid package ID.")
		return
	}

	var req services.PackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body: " + err.Error())
		return
	}

	pkg, err := h.packageService.UpdatePackage(id, req)
	if err != nil {
		if err.Error() == "package not found" {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalServerError(c, "Failed to update package: " + err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Package updated successfully.", pkg)
}

func (h *PackageHandler) DeletePackage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid package ID.")
		return
	}

	if err := h.packageService.DeletePackage(id); err != nil {
		if err.Error() == "package not found" {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalServerError(c, "Failed to delete package: " + err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Package deleted successfully.", nil)
}
