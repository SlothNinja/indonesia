package indonesia

import (
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user"
)

func (g *Game) validatePlayerAction(cu *user.User) error {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	cp := g.CurrentPlayer()
	switch {
	case cp.PerformedAction:
		return sn.NewVError("You have already performed an action.")
	case !g.IsCurrentPlayer(cu):
		return sn.NewVError("Only the current player can perform an action.")
	default:
		return nil
	}
}

func (g *Game) validateAdminAction(cu *user.User) error {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	switch {
	case cu == nil, !cu.Admin:
		return sn.NewVError("Only an admin can perform the selected action.")
	default:
		return nil
	}
}

//type MultiActionID int
//
//const (
//	noMultiAction MultiActionID = iota
//	startedEmpireMA
//	boughtArmiesMA
//	equippedArmyMA
//	placedArmiesMA
//	usedScribeMA
//	selectedWorkerMA
//	placedWorkerMA
//	tradedResourceMA
//	expandEmpireMA
//	builtCityMA
//)
