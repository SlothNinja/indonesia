package indonesia

import (
	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/mlog"
	gtype "github.com/SlothNinja/type"
	"github.com/SlothNinja/user"
	stats "github.com/SlothNinja/user-stats"
	"github.com/gin-gonic/gin"
)

type server struct {
	*datastore.Client
}

func NewClient(dsClient *datastore.Client) server {
	return server{Client: dsClient}
}

func (svr server) addRoutes(prefix string, engine *gin.Engine) *gin.Engine {
	// Game group
	g := engine.Group(prefix + "/game")

	// New
	g.GET("/new",
		user.RequireCurrentUser(),
		gtype.SetTypes(),
		svr.new(prefix),
	)

	// Create
	g.POST("",
		user.RequireCurrentUser(),
		svr.create(prefix),
	)

	// Show
	g.GET("/show/:hid",
		svr.fetch,
		mlog.Get,
		game.SetAdmin(false),
		svr.show(prefix),
	)

	// Undo
	g.POST("/undo/:hid",
		svr.fetch,
		svr.undo(prefix),
	)

	// Finish
	g.POST("/finish/:hid",
		svr.fetch,
		stats.Fetch(user.CurrentFrom),
		svr.finish(prefix),
	)

	// Drop
	g.POST("/drop/:hid",
		user.RequireCurrentUser(),
		svr.fetch,
		svr.drop(prefix),
	)

	// Accept
	g.POST("/accept/:hid",
		user.RequireCurrentUser(),
		svr.fetch,
		svr.accept(prefix),
	)

	// Update
	g.PUT("/show/:hid",
		user.RequireCurrentUser(),
		svr.fetch,
		game.RequireCurrentPlayerOrAdmin(),
		game.SetAdmin(false),
		svr.update(prefix),
	)

	// Add Message
	g.PUT("/show/:hid/addmessage",
		user.RequireCurrentUser(),
		mlog.Get,
		mlog.AddMessage(prefix),
	)

	// Games group
	gs := engine.Group(prefix + "/games")

	// Index
	gs.GET("/:status",
		gtype.SetTypes(),
		svr.index(prefix),
	)

	// JSON Data for Index
	gs.POST("/:status/json",
		gtype.SetTypes(),
		game.GetFiltered(gtype.Indonesia),
		svr.jsonIndexAction(prefix),
	)

	// Admin group
	admin := g.Group("/admin", user.RequireAdmin)

	// Admin
	admin.GET("/:hid",
		svr.fetch,
		mlog.Get,
		game.SetAdmin(true),
		svr.show(prefix),
	)

	// Admin Update
	admin.POST("/:hid",
		svr.fetch,
		game.SetAdmin(true),
		svr.update(prefix),
	)

	admin.PUT("/:hid",
		svr.fetch,
		game.SetAdmin(true),
		svr.update(prefix),
	)

	return engine

}
