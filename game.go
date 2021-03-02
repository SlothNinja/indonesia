package indonesia

import (
	"encoding/gob"
	"errors"
	"html/template"
	"math/rand"
	"time"

	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
	gtype "github.com/SlothNinja/type"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.Register(new(setupEntry))
	gob.Register(new(startEntry))
}

func (client *Client) register(t gtype.Type) *Client {
	gob.Register(new(Game))
	game.Register(t, newGamer, PhaseNames, nil)
	return client.addRoutes(t.Prefix())
}

var ErrMustBeGame = errors.New("Resource must have type *Game.")

const NoPlayerID = game.NoPlayerID

type Game struct {
	*game.Header
	*State
}

type State struct {
	Playerers          game.Playerers
	Log                GameLog
	Era                Era
	AvailableDeeds     Deeds
	Areas              Areas
	CityStones         []int
	Merger             *Merger
	SiapFajiMerger     *SiapFajiMerger
	OverrideDeliveries int
	Version            int
	*TempData
}

// Non-persistent values
// They are cached but ignored by datastore
type TempData struct {
	SelectedSlot             int
	SelectedAreaID           AreaID
	SelectedArea2ID          AreaID
	SelectedGoodsAreaID      AreaID
	SelectedShippingProvince Province
	OldSelectedAreaID        AreaID
	SelectedShipperIndex     int
	SelectedShipper2Index    int
	SelectedCardIndex        int
	SelectedPlayerID         int
	SelectedTechnology       Technology
	SelectedDeedIndex        int
	ShippingCompanyOwnerID   int
	ShippingCompanySlot      int
	ShipsUsed                int
	Expansions               int
	RequiredExpansions       int
	RequiredDeliveries       int
	ProposedPath             flowMatrix
	CustomPath               flowMatrix
	ShipperIncomeMap         ShipperIncomeMap
	Admin                    bool
	AdminAction              string
}

type Era int

const (
	NoEra Era = iota
	EraA
	EraB
	EraC
)

func (e Era) String() string {
	switch e {
	case EraA:
		return "a"
	case EraB:
		return "b"
	case EraC:
		return "c"
	default:
		return ""
	}
}

type ShipType int
type ShipTypes []ShipType

const (
	NoShipType ShipType = iota
	RedShipA
	YellowShipA
	BlueShipA
	RedShipB
	YellowShipB
	BlueShipB
)

var validShipTypes = ShipTypes{
	RedShipA,
	YellowShipA,
	BlueShipA,
	RedShipB,
	YellowShipB,
	BlueShipB,
}

var shipTypeStringMap = map[ShipType]string{
	NoShipType:  "None",
	RedShipA:    "Red Ship A",
	YellowShipA: "Yellow Ship A",
	BlueShipA:   "Blue Ship A",
	RedShipB:    "Red Ship B",
	YellowShipB: "Yellow Ship B",
	BlueShipB:   "Blue Ship B",
}

func (s ShipType) String() string {
	return shipTypeStringMap[s]
}

func (s ShipType) IDString() string {
	return restful.IDString(s.String())
}

func (g *Game) GetPlayerers() game.Playerers {
	return g.Playerers
}

func (g *Game) Players() (players Players) {
	ps := g.GetPlayerers()
	length := len(ps)
	if length > 0 {
		players = make(Players, length)
		for i, p := range ps {
			players[i] = p.(*Player)
		}
	}
	return
}

func (g *Game) setPlayers(players Players) {
	length := len(players)
	if length > 0 {
		ps := make(game.Playerers, length)
		for i, p := range players {
			ps[i] = p
		}
		g.Playerers = ps
	}
}

type Games []*Game

func (client *Client) Start(c *gin.Context, g *Game) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g.Status = game.Running
	g.Version = 2
	g.setupPhase(c)
	return client.start(c, g)
}

func (g *Game) addNewPlayers() {
	for _ = range g.UserIDS {
		g.addNewPlayer()
	}
}

func (g *Game) setupPhase(c *gin.Context) {
	g.Turn = 0
	g.Phase = Setup
	g.CityStones = []int{12, 8, 3}
	g.addNewPlayers()
	g.RandomTurnOrder()
	g.dealCityCards()
	g.createAreas()
	for _, p := range g.Players() {
		g.newSetupEntryFor(p)
	}
	g.beginningOfPhaseReset()
}

func (g *Game) getAvailableShipType() ShipType {
	for _, shipType := range validShipTypes {
		found := false
		for _, shippingCompany := range g.ShippingCompanies() {
			if shippingCompany.ShipType == shipType {
				found = true
				break
			}
		}
		if !found {
			return shipType
		}
	}
	return NoShipType
}

type setupEntry struct {
	*Entry
}

func (g *Game) newSetupEntryFor(p *Player) (e *setupEntry) {
	e = new(setupEntry)
	e.Entry = g.newEntryFor(p)
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return
}

func (e *setupEntry) HTML(c *gin.Context) template.HTML {
	g := gameFrom(c)
	return restful.HTML("%s received 100 rupiah and 3 city cards.", g.NameByPID(e.PlayerID))
}

func (client *Client) start(c *gin.Context, g *Game) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g.Phase = StartGame
	g.newStartEntry()
	_, err := client.startNewEra(c, g)
	return err
}

type startEntry struct {
	*Entry
}

func (g *Game) newStartEntry() *startEntry {
	e := new(startEntry)
	e.Entry = g.newEntry()
	g.Log = append(g.Log, e)
	return e
}

func (e *startEntry) HTML(c *gin.Context) template.HTML {
	g := gameFrom(c)
	names := make([]string, g.NumPlayers)
	for i, p := range g.Players() {
		names[i] = g.NameFor(p)
	}
	return restful.HTML("Good luck %s.  Have fun.", restful.ToSentence(names))
}

func (g *Game) setCurrentPlayers(players ...*Player) {
	var playerers game.Playerers

	switch length := len(players); {
	case length == 0:
		playerers = nil
	case length == 1:
		playerers = game.Playerers{players[0]}
	default:
		playerers = make(game.Playerers, length)
		for i, player := range players {
			playerers[i] = player
		}
	}
	g.SetCurrentPlayerers(playerers...)
}

func (g *Game) PlayerByID(id int) (player *Player) {
	if p := g.PlayererByID(id); p != nil {
		player = p.(*Player)
	}
	return
}

func (g *Game) PlayerBySID(sid string) (player *Player) {
	if p := g.Header.PlayerBySID(sid); p != nil {
		player = p.(*Player)
	}
	return
}

func (g *Game) PlayerByUserID(id int64) (player *Player) {
	if p := g.PlayererByUserID(id); p != nil {
		player = p.(*Player)
	}
	return
}

func (g *Game) PlayerByIndex(index int) (player *Player) {
	if p := g.PlayererByIndex(index); p != nil {
		player = p.(*Player)
	}
	return
}

func (g *Game) undoAction(c *gin.Context, cu *user.User) (string, game.ActionType, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	tmpl, err := g.undoRedoReset(c, cu, "%s undid action.")
	if err != nil {
		return tmpl, game.None, err
	}
	return tmpl, game.Undo, nil
}

func (g Game) redoAction(c *gin.Context, cu *user.User) (string, game.ActionType, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	tmpl, err := g.undoRedoReset(c, cu, "%s redid action.")
	if err != nil {
		return tmpl, game.None, err
	}
	return tmpl, game.Redo, nil
}

func (g *Game) resetTurn(c *gin.Context, cu *user.User) (string, game.ActionType, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	tmpl, err := g.undoRedoReset(c, cu, "%s reset turn.")
	if err != nil {
		return tmpl, game.None, err
	}
	return tmpl, game.Reset, nil
}

func (g *Game) undoRedoReset(c *gin.Context, cu *user.User, fmt string) (tmpl string, err error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp := g.CurrentPlayer()
	if !g.IsCurrentPlayer(cu) {
		err = sn.NewVError("Only the current player may perform this action.")
	}

	restful.AddNoticef(c, fmt, g.NameFor(cp))
	return
}

func (g *Game) CurrentPlayer() *Player {
	p := g.CurrentPlayerer()
	if p != nil {
		return p.(*Player)
	}
	return nil
}

func (g *Game) adminHeader(c *gin.Context, cu *user.User) (string, game.ActionType, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	h := struct {
		Title         string           `form:"title"`
		Turn          int              `form:"turn" binding:"min=0"`
		Phase         game.Phase       `form:"phase" binding:"min=0"`
		SubPhase      game.SubPhase    `form:"sub-phase" binding:"min=0"`
		Round         int              `form:"round" binding:"min=0"`
		NumPlayers    int              `form:"num-players" binding"min=0,max=5"`
		Password      string           `form:"password"`
		CreatorID     int64            `form:"creator-id"`
		CreatorSID    string           `form:"creator-sid"`
		CreatorName   string           `form:"creator-name"`
		UserIDS       []int64          `form:"user-ids"`
		UserSIDS      []string         `form:"user-sids"`
		UserNames     []string         `form:"user-names"`
		UserEmails    []string         `form:"user-emails`
		OrderIDS      game.UserIndices `form:"order-ids"`
		CPUserIndices game.UserIndices `form:"cp-user-indices"`
		WinnerIDS     game.UserIndices `form:"winner-ids"`
		Status        game.Status      `form:"status"`
		Progress      string           `form:"progress"`
		Options       []string         `form:"options"`
		OptString     string           `form:"opt-string"`
		CreatedAt     time.Time        `form:"created-at"`
		UpdatedAt     time.Time        `form:"updated-at"`
	}{}

	err := c.ShouldBind(&h)
	if err != nil {
		return "", game.None, err
	}

	g.Title = h.Title
	g.Turn = h.Turn
	g.Phase = h.Phase
	g.SubPhase = h.SubPhase
	g.Round = h.Round
	g.NumPlayers = h.NumPlayers
	g.Password = h.Password
	g.CreatorID = h.CreatorID
	g.UserIDS = h.UserIDS
	g.OrderIDS = h.OrderIDS
	g.CPUserIndices = h.CPUserIndices
	g.WinnerIDS = h.WinnerIDS
	g.Status = h.Status
	return "", game.Save, nil
}

func (g *Game) adminCities(c *gin.Context, cu *user.User) (string, game.ActionType, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	form := struct {
		CityStones []int `form:"city-stones"`
	}{}

	err := c.ShouldBind(&form)
	if err != nil {
		return "", game.None, err
	}
	// if err = restful.BindWith(c, form, binding.FormPost); err != nil {
	// 	act = game.None
	// 	return
	// }

	for i, s := range form.CityStones {
		g.CityStones[i] = s
	}

	// act = game.Save
	return "", game.Save, nil
}

//func adminHeader(g *Game, form url.Values) (string, game.ActionType, error) {
//	if err := g.adminUpdateHeader(headerValues); err != nil {
//		return "indonesia/flash_notice", game.None, err
//	}
//
//	return "", game.Save, nil
//}
//
//func (g *Game) adminUpdateHeader(ss sslice) (err error) {
//	if err = g.validateAdminAction(ctx); err != nil {
//		return err
//	}
//
//	//g.debugf("Values: %#v", values)
//	mergerRemove, siapFajiMergerRemove := false, false
//	//	addDeedIndex, removeDeedIndex := -1, -1
//	for key := range values {
//		if key == "MergerRemove" {
//			if value := values.Get(key); value == "true" {
//				mergerRemove = true
//			}
//		}
//		if key == "MergerRemove" {
//			if value := values.Get(key); value == "true" {
//				siapFajiMergerRemove = true
//			}
//		}
//		if key == "AddAvailableDeed" {
//			if k := values.Get(key); k != "none" {
//				if d := g.Deeds().get(k); d != nil {
//					g.AvailableDeeds = append(g.AvailableDeeds, d)
//				}
//			}
//		}
//		if key == "RemoveAvailableDeed" {
//			if k := values.Get(key); k != "none" {
//				if d := g.AvailableDeeds.get(k); d != nil {
//					g.AvailableDeeds = g.AvailableDeeds.remove(d)
//				}
//			}
//		}
//		if !ss.include(key) {
//			delete(values, key)
//		}
//	}
//
//	schema.RegisterConverter(game.Phase(0), convertPhase)
//	schema.RegisterConverter(game.SubPhase(0), convertSubPhase)
//	schema.RegisterConverter(game.Status(0), convertStatus)
//	//	game.RegisterDBIDConverter()
//	if err := schema.Decode(g, values); err != nil {
//		return err
//	}
//	if mergerRemove {
//		g.Merger = nil
//	}
//	if siapFajiMergerRemove {
//		g.SiapFajiMerger = nil
//	}
//	//	if addDeedIndex != -1 {
//	//		g.AvailableDeeds = append(g.AvailableDeeds, g.Deeds()[addDeedIndex])
//	//	}
//	//	if removeDeedIndex != -1 {
//	//		g.AvailableDeeds = g.AvailableDeeds.removeAt(removeDeedIndex)
//	//	}
//	return nil
//}
//
//func convertPhase(value string) reflect.Value {
//	if v, err := strconv.ParseInt(value, 10, 0); err == nil {
//		return reflect.ValueOf(game.Phase(v))
//	}
//	return reflect.Value{}
//}
//
//func convertSubPhase(value string) reflect.Value {
//	if v, err := strconv.ParseInt(value, 10, 0); err == nil {
//		return reflect.ValueOf(game.SubPhase(v))
//	}
//	return reflect.Value{}
//}
//
//func convertStatus(value string) reflect.Value {
//	if v, err := strconv.ParseInt(value, 10, 0); err == nil {
//		return reflect.ValueOf(game.Status(v))
//	}
//	return reflect.Value{}
//}

func (g *Game) SelectedPlayer() *Player {
	return g.PlayerByID(g.SelectedPlayerID)
}

func (g *Game) setSelectedPlayer(p *Player) {
	if p != nil {
		g.SelectedPlayerID = p.ID()
	} else {
		g.SelectedPlayerID = NoPlayerID
	}
}

func min(ints ...int) int {
	if len(ints) <= 0 {
		return 0
	}

	min := ints[0]
	for _, i := range ints {
		if i < min {
			min = i
		}
	}
	return min
}

func max(ints ...int) int {
	if len(ints) <= 0 {
		return 0
	}

	max := ints[0]
	for _, i := range ints {
		if i > max {
			max = i
		}
	}
	return max
}

func (g *Game) RandomTurnOrder() {
	rand.Shuffle(len(g.Playerers), func(i, j int) {
		g.Playerers[i], g.Playerers[j] = g.Playerers[j], g.Playerers[i]
	})

	g.SetCurrentPlayerers(g.Playerers[0])

	g.OrderIDS = make(game.UserIndices, len(g.Playerers))
	for i, p := range g.Playerers {
		g.OrderIDS[i] = p.ID()
	}
}
