package api

import (
	"visuilizer/anilist"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	Client *anilist.Client
}

func NewServer(client *anilist.Client) *Server {
	return &Server{Client: client}
}

func (s *Server) Router() *gin.Engine {
	r := gin.Default()

	corsCfg := cors.DefaultConfig()
	corsCfg.AllowOrigins = []string{"http://localhost:5173"}
	corsCfg.AllowMethods = []string{"GET", "POST", "DELETE", "PATCH", "OPTIONS"}
	r.Use(cors.New(corsCfg))

	r.GET("/health", s.HandleHealth)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/media/:id", s.HandleGetMedia)
		v1.GET("/franchise/:id", s.HandleGetFranchise)
	}

	return r
}
