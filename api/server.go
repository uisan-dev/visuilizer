package api

import (
	"visuilizer/anilist"
	"visuilizer/importer"
	"visuilizer/store"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	Client   *anilist.Client
	Store    *store.Store
	Importer *importer.Importer
}

func NewServer(client *anilist.Client, store *store.Store) *Server {
	return &Server{Client: client, Store: store, Importer: importer.NewImporter(client, store)}
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
		v1.POST("/franchise/:id/layouts", s.HandleSaveLayout)
		v1.GET("/franchise/:id/layouts", s.HandleListLayouts)

		v1.POST("/import/:id", s.HandleImport)
		v1.GET("/import/:id", s.HandleImportStatus)

		v1.GET("/layouts/:id", s.HandleGetLayout)
		v1.GET("/layouts/:id/svg", s.HandleGetLayoutSVG)

		v1.GET("/graph/:id", s.HandleGetGraph)
	}

	return r
}
