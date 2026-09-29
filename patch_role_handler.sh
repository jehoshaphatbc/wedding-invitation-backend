#!/bin/bash
cat << 'INNER_EOF' > /tmp/role_handler_patch.go

func (h *RoleHandler) GetTrashedRoles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	search := c.Query("search")
	sort := c.Query("sort")
	order := c.Query("order")

	roles, total, err := h.roleService.GetTrashedRoles(page, perPage, search, sort, order)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve trashed roles")
		return
	}

	meta := gin.H{
		"page":     page,
		"per_page": perPage,
		"total":    total,
	}

	response.SuccessWithMeta(c, http.StatusOK, "Trashed roles retrieved successfully.", roles, meta)
}

func (h *RoleHandler) RestoreRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid role ID.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	if err := h.roleService.RestoreRole(id, ip, userAgent); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Role restored successfully.", nil)
}

func (h *RoleHandler) ForceDeleteRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid role ID.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	if err := h.roleService.ForceDeleteRole(id, ip, userAgent); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Role permanently deleted.", nil)
}
INNER_EOF
cat /tmp/role_handler_patch.go >> pkg/handlers/role_handler.go
