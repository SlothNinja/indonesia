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
	"github.com/patrickmn/go-cache"
)

type Client struct {
	DS     *datastore.Client
	User   user.Client
	Stats  stats.Client
	MLog   mlog.Client
	Game   game.Client
	Rating rating.Client
	Cache  *cache.Cache
}

func NewClient(dsClient *datastore.Client, userClient user.Client, mcache *cache.Cache) Client {
	return Client{
		DS:     dsClient,
		User:   userClient,
		Stats:  stats.NewClient(userClient, dsClient),
		MLog:   mlog.NewClient(userClient, dsClient),
		Game:   game.NewClient(userClient, dsClient),
		Rating: rating.NewClient(userClient, dsClient),
		Cache:  mcache,
	}
}

func (client Client) addRoutes(prefix string, engine *gin.Engine) *gin.Engine {
	// Game group
	g := engine.Group(prefix + "/game")

	// New
	g.GET("/new",
		client.new(prefix),
	)

	// Create
	g.POST("",
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
		client.Stats.Fetch,
		client.finish(prefix),
	)

	// Drop
	g.POST("/drop/:hid",
		client.fetch,
		client.drop(prefix),
	)

	// Accept
	g.POST("/accept/:hid",
		client.fetch,
		client.accept(prefix),
	)

	// Update
	g.PUT("/show/:hid",
		client.fetch,
		game.SetAdmin(false),
		client.update(prefix),
	)

	// Add Message
	g.PUT("/show/:hid/addmessage",
		client.MLog.Get,
		client.MLog.AddMessage(prefix),
	)

	// Games group
	gs := engine.Group(prefix + "/games")

	// Index
	gs.GET("/:status",
		client.index(prefix),
	)

	// JSON Data for Index
	gs.POST("/:status/json",
		client.Game.GetFiltered(gtype.Indonesia),
		client.jsonIndexAction(prefix),
	)

	// Admin group
	admin := g.Group("/admin")

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
