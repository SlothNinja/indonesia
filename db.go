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

func (client Client) init(c *gin.Context, g *Game) error {
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

//func (g *Game) Save(c chan<- datastore.Property) error {
//	// Time stamp
//	t := time.Now()
//	if g.CreatedAt.IsZero() {
//		g.CreatedAt = t
//	}
//	g.UpdatedAt = t
//
//	// Set turn order in header
//	g.OrderIDS = make(game.UserIndices, len(g.Players()))
//	for i, p := range g.Players() {
//		g.OrderIDS[i] = p.ID()
//	}
//
//	// Clear TempData
//	g.TempData = nil
//
//	// Encode and save game state in header
//	if saved, err := codec.Encode(g.State); err != nil {
//		return err
//	} else {
//		g.SavedState = saved
//		return datastore.SaveStruct(g.GetHeader(), c)
//	}
//}

// func (g *Game) Load(props datastore.PropertyMap) error {
// 	h := g.GetHeader()
// 	if err := datastore.GetPLS(h).Load(props); err != nil {
// 		return err
// 	}
//
// 	g.State = newState()
//
// 	if err := codec.Decode(g.State, g.SavedState); err != nil {
// 		return err
// 	}
// 	return g.init(g.CTX())
// }
//
// func (g *Game) Save(withMeta bool) (datastore.PropertyMap, error) {
// 	g.OrderIDS = make(game.UserIndices, len(g.Players()))
// 	for i, p := range g.Players() {
// 		g.OrderIDS[i] = p.ID()
// 	}
//
// 	// Clear TempData
// 	g.TempData = nil
//
// 	if saved, err := codec.Encode(g.State); err != nil {
// 		return nil, err
// 	} else {
// 		g.SavedState = saved
// 		return datastore.GetPLS(g).Save(withMeta)
// 	}
// }

//func (g *Game) Load(c <-chan datastore.Property) error {
//	h := g.GetHeader()
//	if err := datastore.LoadStruct(h, c); err != nil {
//		return err
//	}
//	if err := codec.Decode(g.State, g.SavedState); err != nil {
//		return err
//	}
//	return g.init(g.CTX())
//}

func (client Client) AfterCache(c *gin.Context, g *Game) error {
	return client.init(c, g)
}

func copyGame(g Game) *Game {
	g1 := g
	return &g1
}
