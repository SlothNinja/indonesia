package indonesia

import (
	"strconv"
	"strings"

	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
)

func (g *Game) selectArea(c *gin.Context, cu *user.User) (string, game.ActionType, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	err := g.validateSelectArea(c, cu)
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
			tmpl, err := g.playCard(c, cu)
			return tmpl, game.Cache, err
		case cp.CanPlaceCity():
			tmpl, err := g.placeCity(c, cu)
			return tmpl, game.Cache, err
		case cp.CanAcquireCompany():
			tmpl, err := g.acquireCompany(c, cu)
			return tmpl, game.Cache, err
		case cp.CanResearch():
			tmpl, err := g.conductResearch(c, cu)
			return tmpl, game.Cache, err
		case cp.CanSelectCompanyToOperate():
			tmpl, err := g.selectCompany(c, cu)
			return tmpl, game.Cache, err
		case cp.CanSelectGood():
			tmpl, err := g.selectGood(c, cu)
			return tmpl, game.Cache, err
		case cp.CanSelectShip():
			tmpl, err := g.selectShip(c, cu)
			return tmpl, game.Cache, err
		case cp.CanSelectCityOrShip():
			tmpl, err := g.selectCityOrShip(c, cu)
			return tmpl, game.Cache, err
		case cp.CanExpandProduction():
			tmpl, err := g.expandProduction(c, cu)
			return tmpl, game.Cache, err
		case cp.canExpandShipping():
			tmpl, err := g.expandShipping(c, cu)
			return tmpl, game.Cache, err
		case cp.CanAnnounceMerger():
			tmpl, err := g.selectCompany1(c, cu)
			return tmpl, game.Cache, err
		case cp.CanAnnounceSecondCompany():
			tmpl, err := g.selectCompany2(c, cu)
			return tmpl, game.Cache, err
		case cp.canPlaceInitialProduct():
			tmpl, err := g.placeInitialProduct(c, cu)
			return tmpl, game.Cache, err
		case cp.canPlaceInitialShip():
			tmpl, err := g.placeInitialShip(c, cu)
			return tmpl, game.Cache, err
		case cp.CanCreateSiapFaji():
			tmpl, err := g.removeRiceSpice(c, cu)
			return tmpl, game.Cache, err
		default:
			return "indonesia/flash_notice", game.None, sn.NewVError("Can't find action for selection.")
		}
	}
}

func (g *Game) validateSelectArea(c *gin.Context, cu *user.User) error {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if !g.IsCurrentPlayer(cu) {
		return sn.NewVError("Only the current player can perform an action.")
	}

	areaID := c.PostForm("area")

	splits := strings.Split(areaID, "-")
	switch {
	case splits[0] == "admin" && splits[1] == "area":
		g.AdminAction = "admin-area"
		id, err := strconv.Atoi(splits[2])
		if err != nil {
			return err
		}
		g.SelectedAreaID = AreaID(id)
		return nil
	case splits[0] == "admin" && splits[1] == "player":
		g.AdminAction = "admin-player"
		id, err := strconv.Atoi(splits[2])
		if err != nil {
			return err
		}
		g.SelectedPlayerID = id
		return nil
	case splits[0] == "admin" && splits[1] == "company":
		g.AdminAction = "admin-company"

		id, err := strconv.Atoi(splits[2])
		if err != nil {
			return err
		}

		slot, err := strconv.Atoi(splits[3])
		if err != nil {
			return err
		}

		g.SelectedPlayerID = id
		g.SelectedSlot = slot
		return nil
	case splits[0] == "admin":
		g.AdminAction = areaID
		return nil
	case splits[0] == "card":
		i, err := strconv.Atoi(splits[1])
		if err != nil {
			return err
		}
		g.SelectedCardIndex = i
		return nil
	case splits[0] == "available":
		i, err := strconv.Atoi(splits[2])
		if err != nil {
			return err
		}
		g.SelectedDeedIndex = i
		return nil
	case splits[0] == "area":
		id, err := strconv.Atoi(splits[1])
		if err != nil {
			return err
		}
		g.SelectedAreaID = AreaID(id)
		return nil
	case splits[0] == "research":
		i, err := strconv.Atoi(splits[1])
		if err != nil {
			return err
		}
		g.SelectedTechnology = Technology(i)
		return nil
	case splits[0] == "company":
		i, err := strconv.Atoi(splits[1])
		if err != nil {
			return err
		}
		g.SelectedSlot = i
		g.setSelectedPlayer(g.CurrentPlayer())
		return nil
	case splits[0] == "ship":
		i, err := strconv.Atoi(splits[1])
		if err != nil {
			return err
		}

		j, err := strconv.Atoi(splits[2])
		if err != nil {
			return err
		}
		g.SelectedArea2ID = AreaID(i)
		g.SelectedShipperIndex = j
		return nil
	case splits[0] == "city":
		i, err := strconv.Atoi(splits[1])
		if err != nil {
			return err
		}
		g.SelectedArea2ID = AreaID(i)
		return nil
	case splits[0] == "player":
		id, err := strconv.Atoi(splits[1])
		if err != nil {
			return err
		}

		slot, err := strconv.Atoi(splits[3])
		if err != nil {
			return err
		}
		g.SelectedPlayerID = id
		g.SelectedSlot = slot
		return nil
	default:
		return sn.NewVError("Unable to determine selection.")
	}
}
