#!/bin/bash
cat << 'INNER_EOF' > /tmp/user_handler_patch.go

func (h *UserHandler) GetTrashedUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	search := c.Query("search")
	sort := c.Query("sort")
	order := c.Query("order")

	users, total, err := h.userService.GetTrashedUsers(page, perPage, search, sort, order)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve trashed users")
		return
	}

	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, models.ToUserResponse(&user))
	}

	meta := gin.H{
		"page":     page,
		"per_page": perPage,
		"total":    total,
	}

	response.SuccessWithMeta(c, http.StatusOK, "Trashed users retrieved successfully.", userResponses, meta)
}

func (h *UserHandler) RestoreUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	if err := h.userService.RestoreUser(id, ip, userAgent); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "User restored successfully.", nil)
}

func (h *UserHandler) ForceDeleteUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID.")
		return
	}

	ip := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	if err := h.userService.ForceDeleteUser(id, ip, userAgent); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "User permanently deleted.", nil)
}
INNER_EOF
cat /tmp/user_handler_patch.go >> pkg/handlers/user_handler.go
