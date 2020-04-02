package indonesia

import (
	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/mlog"
	"github.com/SlothNinja/rating"
	gtype "github.com/SlothNinja/type"
	"github.com/SlothNinja/user"
	stats "github.com/SlothNinja/user-stats"
	"github.com/gin-gonic/gin"
)

type Client struct {
	*datastore.Client
	Stats  stats.Client
	MLog   mlog.Client
	Game   game.Client
	Rating rating.Client
}

func NewClient(dsClient *datastore.Client) Client {
	return Client{
		Client: dsClient,
		Stats:  stats.NewClient(dsClient),
		MLog:   mlog.NewClient(dsClient),
		Game:   game.NewClient(dsClient),
		Rating: rating.NewClient(dsClient),
	}
}

func (client Client) addRoutes(prefix string, engine *gin.Engine) *gin.Engine {
	// Game group
	g := engine.Group(prefix + "/game")

	// New
	g.GET("/new",
		user.RequireCurrentUser(),
		gtype.SetTypes(),
		client.new(prefix),
	)

	// Create
	g.POST("",
		user.RequireCurrentUser(),
		client.create(prefix),
	)

	// Show
	g.GET("/show/:hid",
		client.fetch,
		client.MLog.Get,
		game.SetAdmin(false),
		client.show(prefix),
	)

	// Undo
	g.POST("/undo/:hid",
		client.fetch,
		client.undo(prefix),
	)

	// Finish
	g.POST("/finish/:hid",
		client.fetch,
		client.Stats.Fetch(user.CurrentFrom),
		client.finish(prefix),
	)

	// Drop
	g.POST("/drop/:hid",
		user.RequireCurrentUser(),
		client.fetch,
		client.drop(prefix),
	)

	// Accept
	g.POST("/accept/:hid",
		user.RequireCurrentUser(),
		client.fetch,
		client.accept(prefix),
	)

	// Update
	g.PUT("/show/:hid",
		user.RequireCurrentUser(),
		client.fetch,
		game.RequireCurrentPlayerOrAdmin(),
		game.SetAdmin(false),
		client.update(prefix),
	)

	// Add Message
	g.PUT("/show/:hid/addmessage",
		user.RequireCurrentUser(),
		client.MLog.Get,
		client.MLog.AddMessage(prefix),
	)

	// Games group
	gs := engine.Group(prefix + "/games")

	// Index
	gs.GET("/:status",
		gtype.SetTypes(),
		client.index(prefix),
	)

	// JSON Data for Index
	gs.POST("/:status/json",
		gtype.SetTypes(),
		client.Game.GetFiltered(gtype.Indonesia),
		client.jsonIndexAction(prefix),
	)

	// Admin group
	admin := g.Group("/admin", user.RequireAdmin)

	// Admin
	admin.GET("/:hid",
		client.fetch,
		client.MLog.Get,
		game.SetAdmin(true),
		client.show(prefix),
	)

	// Admin Update
	admin.POST("/:hid",
		client.fetch,
		game.SetAdmin(true),
		client.update(prefix),
	)

	admin.PUT("/:hid",
		client.fetch,
		game.SetAdmin(true),
		client.update(prefix),
	)

	return engine

}
