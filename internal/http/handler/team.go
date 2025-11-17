package handler

import "github.com/gin-gonic/gin"

type TeamHandler struct {
}

func NewTeamHandler() *TeamHandler {
	return &TeamHandler{}
}

func (h *TeamHandler) AddTeam(c *gin.Context) {

}

func (h *TeamHandler) GetTeam(c *gin.Context) {

}
