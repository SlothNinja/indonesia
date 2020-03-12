package indonesia

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/codec"
	"github.com/SlothNinja/color"
	"github.com/SlothNinja/contest"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/mlog"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
	gtype "github.com/SlothNinja/type"
	"github.com/SlothNinja/user"
	stats "github.com/SlothNinja/user-stats"
	"github.com/gin-gonic/gin"
	"google.golang.org/appengine"
	"google.golang.org/appengine/memcache"
)

const (
	gameKey   = "Game"
	homePath  = "/"
	jsonKey   = "JSON"
	statusKey = "Status"
	hParam    = "hid"
)

func gameFrom(c *gin.Context) (g *Game) {
	g, _ = c.Value(gameKey).(*Game)
	return
}

func withGame(c *gin.Context, g *Game) *gin.Context {
	c.Set(gameKey, g)
	return c
}

func jsonFrom(c *gin.Context) (g *Game) {
	g, _ = c.Value(jsonKey).(*Game)
	return
}

func withJSON(c *gin.Context, g *Game) *gin.Context {
	c.Set(jsonKey, g)
	return c
}

//type Action func(*Game, url.Values) (string, game.ActionType, error)
//
//var actionMap = map[string]Action{
//	"select-area":              selectArea,
//	"select-hull-player":       selectHullPlayer,
//	"turn-order-bid":           placeTurnOrderBid,
//	"stop-expanding":           stopExpanding,
//	"accept-proposed-flow":     acceptProposedFlow,
//	"city-growth":              cityGrowth,
//	"pass":                     pass,
//	"merger-bid":               mergerBid,
//	"undo":                     undoAction,
//	"redo":                     redoAction,
//	"reset":                    resetTurn,
//	"finish":                   finishTurn,
//	"admin-header":             adminHeader,
//	"admin-area":               adminArea,
//	"admin-patch":              adminPatch,
//	"admin-player":             adminPlayer,
//	"admin-company":            adminCompany,
//	"admin-player-new-company": adminPlayerNewCompany,
//}

func (g *Game) Update(c *gin.Context) (tmpl string, act game.ActionType, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	switch a := c.PostForm("action"); a {
	case "select-area":
		tmpl, act, err = g.selectArea(c)
	case "select-hull-player":
		tmpl, act, err = g.selectHullPlayer(c)
	case "turn-order-bid":
		tmpl, act, err = g.placeTurnOrderBid(c)
	case "stop-expanding":
		tmpl, act, err = g.stopExpanding(c)
	case "accept-proposed-flow":
		tmpl, act, err = g.acceptProposedFlow(c)
	case "city-growth":
		tmpl, act, err = g.cityGrowth(c)
	case "pass":
		tmpl, act, err = g.pass(c)
	case "merger-bid":
		tmpl, act, err = g.mergerBid(c)
	case "undo":
		tmpl, act, err = g.undoAction(c)
	case "redo":
		tmpl, act, err = g.redoAction(c)
	case "reset":
		tmpl, act, err = g.resetTurn(c)
	//	case "finish":
	//		tmpl, act, err = g.finishTurn(c)
	case "admin-header":
		tmpl, act, err = g.adminHeader(c)
	case "admin-cities":
		tmpl, act, err = g.adminCities(c)
		//	"admin-area":               adminArea,
		//	"admin-patch":              adminPatch,
		//	"admin-player":             adminPlayer,
		//	"admin-company":            adminCompany,
		//	"admin-player-new-company": adminPlayerNewCompany,
	default:
		tmpl, act, err = "indonesia/flash_notice", game.None, sn.NewVError("%v is not a valid action.", a)
	}
	return
}

func Show(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")

		g := gameFrom(c)
		cu := user.CurrentFrom(c)
		c.HTML(http.StatusOK, prefix+"/show", gin.H{
			"Context":    c,
			"VersionID":  appengine.VersionID(appengine.NewContext(c.Request)),
			"CUser":      cu,
			"Game":       g,
			"IsAdmin":    user.IsAdmin(c),
			"Admin":      game.AdminFrom(c),
			"MessageLog": mlog.From(c),
			"ColorMap":   color.MapFrom(c),
			"Notices":    restful.NoticesFrom(c),
			"Errors":     restful.ErrorsFrom(c),
		})
	}
}

func Update(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")

		g := gameFrom(c)
		if g == nil {
			log.Errorf("Controller#Update Game Not Found")
			c.Redirect(http.StatusSeeOther, homePath)
			return
		}
		template, actionType, err := g.Update(c)
		switch {
		case err != nil && sn.IsVError(err):
			restful.AddErrorf(c, "%v", err)
			withJSON(c, g)
		case err != nil:
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, homePath)
			return
		case actionType == game.Cache:
			if err := g.cache(c); err != nil {
				restful.AddErrorf(c, "%v", err)
			}
		case actionType == game.Save:
			if err := g.save(c); err != nil {
				log.Errorf("%s", err)
				restful.AddErrorf(c, "Controller#Update Save Error: %s", err)
				c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param(hParam)))
				return
			}
		case actionType == game.Undo:
			mkey := g.UndoKey(c)
			if err := memcache.Delete(appengine.NewContext(c.Request), mkey); err != nil && err != memcache.ErrCacheMiss {
				log.Errorf("memcache.Delete error: %s", err)
				c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param(hParam)))
				return
			}
		}

		switch jData := jsonFrom(c); {
		case jData != nil && template == "json":
			log.Debugf("jData: %v", jData)
			log.Debugf("template: %v", template)
			c.JSON(http.StatusOK, jData)
		case template == "":
			log.Debugf("template: %v", template)
			c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param(hParam)))
		default:
			log.Debugf("template: %v", template)
			cu := user.CurrentFrom(c)

			d := gin.H{
				"Context":   c,
				"VersionID": appengine.VersionID(appengine.NewContext(c.Request)),
				"CUser":     cu,
				"Game":      g,
				"Admin":     game.AdminFrom(c),
				"IsAdmin":   user.IsAdmin(c),
				"Notices":   restful.NoticesFrom(c),
				"Errors":    restful.ErrorsFrom(c),
			}
			log.Debugf("d: %#v", d)
			c.HTML(http.StatusOK, template, d)
		}
	}
}
func (g *Game) save(c *gin.Context) error {
	dsClient, err := datastore.NewClient(c, "")
	if err != nil {
		return err
	}

	_, err = dsClient.RunInTransaction(c, func(tx *datastore.Transaction) error {
		oldG := New(c, g.ID())
		err = tx.Get(oldG.Key, oldG.Header)
		if err != nil {
			return err
		}

		if oldG.UpdatedAt != g.UpdatedAt {
			return fmt.Errorf("Game state changed unexpectantly.  Try again.")
		}

		err = g.encode(c)
		if err != nil {
			return err
		}

		_, err = tx.Put(g.Key, g.Header)
		if err != nil {
			return err
		}

		err = memcache.Delete(appengine.NewContext(c.Request), g.UndoKey(c))
		if err == memcache.ErrCacheMiss {
			return nil
		}
		return err
	})
	return err
}

func (g *Game) saveWith(c *gin.Context, ks []*datastore.Key, es []interface{}) error {
	dsClient, err := datastore.NewClient(c, "")
	if err != nil {
		return err
	}

	_, err = dsClient.RunInTransaction(c, func(tx *datastore.Transaction) error {
		oldG := New(c, g.ID())
		err := tx.Get(oldG.Key, oldG.Header)
		if err != nil {
			return err
		}

		if oldG.UpdatedAt != g.UpdatedAt {
			return fmt.Errorf("Game state changed unexpectantly.  Try again.")
		}

		err = g.encode(c)
		if err != nil {
			return err
		}

		ks = append(ks, g.Key)
		es = append(es, g.Header)

		_, err = tx.PutMulti(ks, es)
		if err != nil {
			return err
		}

		err = memcache.Delete(appengine.NewContext(c.Request), g.UndoKey(c))
		if err == memcache.ErrCacheMiss {
			return nil
		}
		return err
	})
	return nil
}

func (g *Game) encode(c *gin.Context) (err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	g.TempData = nil
	var encoded []byte
	if encoded, err = codec.Encode(g.State); err != nil {
		return
	}
	g.SavedState = encoded
	g.updateHeader()

	return
}

func (g *Game) cache(c *gin.Context) error {
	item := &memcache.Item{
		Key:        g.UndoKey(c),
		Expiration: time.Minute * 30,
	}
	v, err := codec.Encode(g)
	if err != nil {
		return err
	}
	item.Value = v
	return memcache.Set(appengine.NewContext(c.Request), item)
}

func wrap(s *stats.Stats, cs contest.Contests) ([]*datastore.Key, []interface{}) {
	l := len(cs) + 1
	es := make([]interface{}, l)
	ks := make([]*datastore.Key, l)
	es[0] = s
	ks[0] = s.Key
	for i, c := range cs {
		es[i+1] = c
		ks[i+1] = c.Key
	}
	return ks, es
}

func showPath(prefix, hid string) string {
	return fmt.Sprintf("/%s/game/show/%s", prefix, hid)
}

func recruitingPath(prefix string) string {
	return fmt.Sprintf("/%s/games/recruiting", prefix)
}

func newPath(prefix string) string {
	return fmt.Sprintf("/%s/game/new", prefix)
}

func newGamer(c *gin.Context) game.Gamer {
	return New(c, 0)
}

func Undo(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")
		c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param(hParam)))

		g := gameFrom(c)
		if g == nil {
			log.Errorf("Controller#Update Game Not Found")
			return
		}
		mkey := g.UndoKey(c)
		if err := memcache.Delete(appengine.NewContext(c.Request), mkey); err != nil && err != memcache.ErrCacheMiss {
			log.Errorf("Controller#Undo Error: %s", err)
		}
	}
}
func Index(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")

		gs := game.GamersFrom(c)
		switch status := game.StatusFrom(c); status {
		case game.Recruiting:
			c.HTML(http.StatusOK, "shared/invitation_index", gin.H{
				"Context":   c,
				"VersionID": appengine.VersionID(appengine.NewContext(c.Request)),
				"CUser":     user.CurrentFrom(c),
				"Games":     gs,
				"Type":      gtype.Indonesia.String(),
			})
		default:
			c.HTML(http.StatusOK, "shared/games_index", gin.H{
				"Context":   c,
				"VersionID": appengine.VersionID(appengine.NewContext(c.Request)),
				"CUser":     user.CurrentFrom(c),
				"Games":     gs,
				"Type":      gtype.Indonesia.String(),
				"Status":    status,
			})
		}
	}
}

func NewAction(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")

		g := New(c, 0)
		withGame(c, g)
		if err := g.FromParams(c, gtype.GOT); err != nil {
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		c.HTML(http.StatusOK, prefix+"/new", gin.H{
			"Context":   c,
			"VersionID": appengine.VersionID(appengine.NewContext(c.Request)),
			"CUser":     user.CurrentFrom(c),
			"Game":      g,
		})
	}
}

func Create(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")

		g := New(c, 0)
		withGame(c, g)

		err := g.FromForm(c, g.Type)
		if err != nil {
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}
		err = g.encode(c)
		if err != nil {
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		dsClient, err := datastore.NewClient(c, "")
		if err != nil {
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		ks, err := dsClient.AllocateIDs(c, []*datastore.Key{g.Key})
		if err != nil {
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		k := ks[0]

		_, err = dsClient.RunInTransaction(c, func(tx *datastore.Transaction) error {
			m := mlog.New(k.ID)

			log.Debugf("m.Key: %v", m.Key)
			log.Debugf("g.Key: %v", k)
			ks := []*datastore.Key{m.Key, k}
			es := []interface{}{m, g.Header}
			_, err := tx.PutMulti(ks, es)
			return err
		})
		if err != nil {
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		restful.AddNoticef(c, "<div>%s created.</div>", g.Title)
		c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
	}
}

func Accept(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")
		defer c.Redirect(http.StatusSeeOther, recruitingPath(prefix))

		g := gameFrom(c)
		if g == nil {
			log.Errorf("game not found")
			return
		}

		var (
			start bool
			err   error
		)

		u := user.CurrentFrom(c)
		if start, err = g.Accept(c, u); err == nil && start {
			err = g.Start(c)
		}

		if err == nil {
			err = g.save(c)
		}

		if err == nil && start {
			g.SendTurnNotificationsTo(c, g.CurrentPlayer())
		}

		if err != nil {
			log.Errorf(err.Error())
		}

	}
}

func Drop(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")
		defer c.Redirect(http.StatusSeeOther, recruitingPath(prefix))

		g := gameFrom(c)
		if g == nil {
			log.Errorf("game not found")
			return
		}

		var err error

		u := user.CurrentFrom(c)
		if err = g.Drop(u); err == nil {
			err = g.save(c)
		}

		if err != nil {
			log.Errorf(err.Error())
			restful.AddErrorf(c, err.Error())
		}

	}
}

func Fetch(c *gin.Context) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")
	// create Gamer
	log.Debugf("hid: %v", c.Param("hid"))
	id, err := strconv.ParseInt(c.Param("hid"), 10, 64)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	log.Debugf("id: %v", id)
	g := New(c, id)
	t := g.Type

	switch action := c.PostForm("action"); {
	case action == "reset":
		// pull from memcache/datastore
		// same as undo & !MultiUndo
		fallthrough
	case action == "undo" && !t.MultiUndo():
		// pull from memcache/datastore
		if err := dsGet(c, g); err != nil {
			c.Redirect(http.StatusSeeOther, homePath)
			return
		}
	default:
		if user.CurrentFrom(c) != nil {
			// pull from memcache and return if successful; otherwise pull from datastore
			if err := mcGet(c, g); err == nil {
				return
			}
		}
		log.Debugf("g: %#v", g)
		log.Debugf("k: %v", g.Key)
		if err := dsGet(c, g); err != nil {
			log.Debugf("dsGet error: %v", err)
			c.Redirect(http.StatusSeeOther, homePath)
			return
		}
	}
}

// pull temporary game state from memcache.  Note may be different from value stored in datastore.
func mcGet(c *gin.Context, g *Game) error {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	mkey := g.GetHeader().UndoKey(c)
	item, err := memcache.Get(appengine.NewContext(c.Request), mkey)
	if err != nil {
		return err
	}

	err = codec.Decode(g, item.Value)
	if err != nil {
		return err
	}

	err = g.AfterCache()
	if err != nil {
		return err
	}

	color.WithMap(withGame(c, g), g.ColorMapFor(user.CurrentFrom(c)))
	return nil
}

// pull game state from memcache/datastore.  returned memcache should be same as datastore.
func dsGet(c *gin.Context, g *Game) error {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	dsClient, err := datastore.NewClient(c, "")
	if err != nil {
		log.Errorf(err.Error())
		return err
	}

	switch err := dsClient.Get(c, g.Key, g.Header); {
	case err != nil:
		restful.AddErrorf(c, err.Error())
		return err
	case g == nil:
		err := fmt.Errorf("Unable to get game for id: %v", g.ID)
		restful.AddErrorf(c, err.Error())
		return err
	}

	s := newState()
	if err := codec.Decode(&s, g.SavedState); err != nil {
		restful.AddErrorf(c, err.Error())
		return err
	} else {
		g.State = s
	}

	if err := g.init(c); err != nil {
		log.Debugf("g.init error: %v", err)
		restful.AddErrorf(c, err.Error())
		return err
	}

	cm := g.ColorMapFor(user.CurrentFrom(c))
	log.Debugf("cm: %#v", cm)
	color.WithMap(withGame(c, g), cm)
	return nil
}

func JSON(c *gin.Context) {
	c.JSON(http.StatusOK, gameFrom(c))
}

func JSONIndexAction(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")

		game.JSONIndexAction(c)
	}
}

func (g *Game) updateHeader() {
	switch g.Phase {
	case GameOver:
		g.Progress = g.PhaseName()
	default:
		g.Progress = fmt.Sprintf("<div>Era: %s | Turn: %d</div><div>Phase: %s</div>", g.Era, g.Turn, g.PhaseName())
	}
	if u := g.Creator; u != nil {
		g.CreatorSID = user.GenID(u.GoogleID)
		g.CreatorName = u.Name
	}

	if l := len(g.Users); l > 0 {
		g.UserSIDS = make([]string, l)
		g.UserNames = make([]string, l)
		g.UserEmails = make([]string, l)
		for i, u := range g.Users {
			g.UserSIDS[i] = user.GenID(u.GoogleID)
			g.UserNames[i] = u.Name
			g.UserEmails[i] = u.Email
		}
	}
}
