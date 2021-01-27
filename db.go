package indonesia

import (
	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/game"
	gtype "github.com/SlothNinja/type"
	"github.com/gin-gonic/gin"
)

const kind = "Game"

func New(c *gin.Context, id int64) *Game {
	g := new(Game)
	g.Header = game.NewHeader(c, g, id)
	g.State = newState()
	g.Key.Parent = pk(c)
	g.Type = gtype.Indonesia
	return g
}

const NoShipIndex = -1

func newState() *State {
	return &State{TempData: new(TempData)}
}

func pk(c *gin.Context) *datastore.Key {
	return datastore.NameKey(gtype.Indonesia.SString(), "root", game.GamesRoot(c))
}

func newKey(c *gin.Context, id int64) *datastore.Key {
	return datastore.IDKey(kind, id, pk(c))
}

func (g *Game) NewKey(c *gin.Context, id int64) *datastore.Key {
	return newKey(c, id)
}

func (client *Client) init(c *gin.Context, g *Game) error {
	err := client.Game.AfterLoad(c, g.Header)
	if err != nil {
		return err
	}

	for _, player := range g.Players() {
		player.Init(g)
	}

	g.initAreas()
	if g.Merger != nil {
		g.Merger.g = g
	}

	if g.SiapFajiMerger != nil {
		g.SiapFajiMerger.init(g)
	}
	return nil
}

func (client *Client) AfterCache(c *gin.Context, g *Game) error {
	return client.init(c, g)
}

func copyGame(g Game) *Game {
	g1 := g
	return &g1
}
