package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/k8sgo/internal/models"
)

// @Summary Create a new user environment
// @Description Creates a new Kubernetes environment for a user
// @Tags environments
// @Accept json
// @Produce json
// @Param environment body models.EnvironmentRequest true "Environment Details"
// @Success 201 {object} models.Environment
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /environments [post]
func (s *Server) createEnvironment(c *gin.Context) {
	var req models.EnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request: " + err.Error(),
		})
		return
	}

	// Validate request
	if req.Username == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Username is required",
		})
		return
	}

	// Create the environment
	env, err := s.k8sClient.CreateEnvironment(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to create environment: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, env)
}

// @Summary Get all environments
// @Description Lists all user environments
// @Tags environments
// @Produce json
// @Success 200 {array} models.Environment
// @Failure 500 {object} models.ErrorResponse
// @Router /environments [get]
func (s *Server) listEnvironments(c *gin.Context) {
	envs, err := s.k8sClient.ListEnvironments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to list environments: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, envs)
}

// @Summary Get a specific environment
// @Description Gets a user environment by username
// @Tags environments
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} models.Environment
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /environments/{username} [get]
func (s *Server) getEnvironment(c *gin.Context) {
	username := c.Param("username")
	env, err := s.k8sClient.GetEnvironment(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to get environment: " + err.Error(),
		})
		return
	}

	if env == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Environment not found",
		})
		return
	}

	c.JSON(http.StatusOK, env)
}

// @Summary Delete an environment
// @Description Deletes a user environment
// @Tags environments
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} models.SuccessResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /environments/{username} [delete]
func (s *Server) deleteEnvironment(c *gin.Context) {
	username := c.Param("username")
	err := s.k8sClient.DeleteEnvironment(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to delete environment: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Environment deleted successfully",
	})
}

// @Summary Update an environment
// @Description Updates a user environment
// @Tags environments
// @Accept json
// @Produce json
// @Param username path string true "Username"
// @Param environment body models.EnvironmentRequest true "Environment Details"
// @Success 200 {object} models.Environment
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /environments/{username} [put]
func (s *Server) updateEnvironment(c *gin.Context) {
	username := c.Param("username")
	var req models.EnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request: " + err.Error(),
		})
		return
	}

	// Override username from URL
	req.Username = username

	// Update the environment
	env, err := s.k8sClient.UpdateEnvironment(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to update environment: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, env)
}