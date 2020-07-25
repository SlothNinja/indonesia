package indonesia

import (
	"strconv"
	"strings"

	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn"
	"github.com/gin-gonic/gin"
)

func (g *Game) selectArea(c *gin.Context) (string, game.ActionType, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	err := g.validateSelectArea(c)
	if err != nil {
		return "indonesia/flash_notice", game.None, err
	}

	cp := g.CurrentPlayer()
	switch g.AdminAction {
	case "admin-header":
		return "indonesia/admin/header_dialog", game.Cache, nil
	case "admin-player":
		return "indonesia/admin/player_dialog", game.Cache, nil
	case "admin-area":
		return "indonesia/admin/area_dialog", game.Cache, nil
	case "admin-company":
		return "indonesia/admin/company_dialog", game.Cache, nil
	default:
		switch {
		case cp.CanSelectCard():
			tmpl, err := g.playCard(c)
			return tmpl, game.Cache, err
		case cp.CanPlaceCity():
			tmpl, err := g.placeCity(c)
			return tmpl, game.Cache, err
		case cp.CanAcquireCompany():
			tmpl, err := g.acquireCompany(c)
			return tmpl, game.Cache, err
		case cp.CanResearch():
			tmpl, err := g.conductResearch(c)
			return tmpl, game.Cache, err
		case cp.CanSelectCompanyToOperate():
			tmpl, err := g.selectCompany(c)
			return tmpl, game.Cache, err
		case cp.CanSelectGood():
			tmpl, err := g.selectGood(c)
			return tmpl, game.Cache, err
		case cp.CanSelectShip():
			tmpl, err := g.selectShip(c)
			return tmpl, game.Cache, err
		case cp.CanSelectCityOrShip():
			tmpl, err := g.selectCityOrShip(c)
			return tmpl, game.Cache, err
		case cp.CanExpandProduction():
			tmpl, err := g.expandProduction(c)
			return tmpl, game.Cache, err
		case cp.canExpandShipping():
			tmpl, err := g.expandShipping(c)
			return tmpl, game.Cache, err
		case cp.CanAnnounceMerger():
			tmpl, err := g.selectCompany1(c)
			return tmpl, game.Cache, err
		case cp.CanAnnounceSecondCompany():
			tmpl, err := g.selectCompany2(c)
			return tmpl, game.Cache, err
		case cp.canPlaceInitialProduct():
			tmpl, err := g.placeInitialProduct(c)
			return tmpl, game.Cache, err
		case cp.canPlaceInitialShip():
			tmpl, err := g.placeInitialShip(c)
			return tmpl, game.Cache, err
		case cp.CanCreateSiapFaji():
			tmpl, err := g.removeRiceSpice(c)
			return tmpl, game.Cache, err
		default:
			return "indonesia/flash_notice", game.None, sn.NewVError("Can't find action for selection.")
		}
	}
}

func (g *Game) validateSelectArea(c *gin.Context) (err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if !g.CUserIsCPlayerOrAdmin(c) {
		err = sn.NewVError("Only the current player can perform an action.")
		return
	}

	var i, id, slot int
	areaID := c.PostForm("area")

	switch splits := strings.Split(areaID, "-"); {
	case splits[0] == "admin" && splits[1] == "area":
		g.AdminAction = "admin-area"
		if id, err = strconv.Atoi(splits[2]); err == nil {
			g.SelectedAreaID = AreaID(id)
		}
	case splits[0] == "admin" && splits[1] == "player":
		g.AdminAction = "admin-player"
		if id, err = strconv.Atoi(splits[2]); err == nil {
			g.SelectedPlayerID = id
		}
	case splits[0] == "admin" && splits[1] == "company":
		g.AdminAction = "admin-company"
		if id, err = strconv.Atoi(splits[2]); err == nil {
			g.SelectedPlayerID = id

			if slot, err = strconv.Atoi(splits[3]); err == nil {
				g.SelectedSlot = slot
			}
		}
	case splits[0] == "admin":
		g.AdminAction = areaID
	case splits[0] == "card":
		if i, err = strconv.Atoi(splits[1]); err == nil {
			g.SelectedCardIndex = i
		}
	case splits[0] == "available":
		if i, err = strconv.Atoi(splits[2]); err == nil {
			g.SelectedDeedIndex = i
		}
	case splits[0] == "area":
		if id, err = strconv.Atoi(splits[1]); err == nil {
			g.SelectedAreaID = AreaID(id)
		}
	case splits[0] == "research":
		if i, err = strconv.Atoi(splits[1]); err == nil {
			g.SelectedTechnology = Technology(i)
		}
	case splits[0] == "company":
		if i, err = strconv.Atoi(splits[1]); err == nil {
			g.SelectedSlot = i
			g.setSelectedPlayer(g.CurrentPlayer())
		}
	case splits[0] == "ship":
		if i, err = strconv.Atoi(splits[1]); err == nil {
			g.SelectedArea2ID = AreaID(i)

			if i, err = strconv.Atoi(splits[2]); err == nil {
				g.SelectedShipperIndex = i
			}
		}
	case splits[0] == "city":
		if i, err = strconv.Atoi(splits[1]); err == nil {
			g.SelectedArea2ID = AreaID(i)
		}
	case splits[0] == "player":
		if id, err = strconv.Atoi(splits[1]); err == nil {
			g.SelectedPlayerID = id

			if slot, err = strconv.Atoi(splits[3]); err == nil {
				g.SelectedSlot = slot
			}
		}
	default:
		err = sn.NewVError("Unable to determine selection.")
	}
	return
}
