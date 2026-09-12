package simulator

import (
	"fmt"
	"strings"
)

/*
Action is an interface for players to perform during simulation. The set of actions players perform during simulation by running NextTurn, NextTurnN, Continuous.
*/
type Action interface {
	Name() string                                       // must be unique
	Condition(pl *Player, s *Simulator) bool            // if false, action will be skipped on turn
	Exec(pl *Player, s *Simulator) error                // executes action if condition is true
	Attributes(pl *Player, s *Simulator) map[string]any // returns an arbitrary attributes map, use this to simulate application data tied to player actions
}

type PlayerAction struct {
	Turn       int64
	ActionName string
	PlayerID   int
	Attributes map[string]any
}

func NewPlayerAction(pl *Player, ac Action, s *Simulator) PlayerAction {
	return PlayerAction{
		Turn:       s.turn,
		ActionName: ac.Name(),
		PlayerID:   pl.id,
		Attributes: ac.Attributes(pl, s),
	}
}

func (pa PlayerAction) String() string {
	return strings.Join(
		[]string{
			fmt.Sprintf("action=%v", pa.ActionName),
			fmt.Sprintf("turn=%v", pa.Turn),
			fmt.Sprintf("player.id=%v", pa.PlayerID),
			fmt.Sprintf("attributes=%v", pa.Attributes),
		},
		"  ",
	)
}

func init() {
	actionMap = map[string]Action{}
	AddAction(OnboardAction{})
	AddAction(LoginAction{})
	AddAction(LogoutAction{})
}

var (
	actionMap   map[string]Action
	actionSlice []Action
)

func AddAction(ac Action) error {
	if _, exists := actionMap[ac.Name()]; exists {
		return fmt.Errorf("duplicate action '%v'", ac.Name())
	}
	actionMap[ac.Name()] = ac
	actionSlice = append(actionSlice, ac)
	return nil
}

/*
Core Actions
*/
type OnboardAction struct{}

func (ac OnboardAction) Name() string { return "onboarding" }
func (ac OnboardAction) Condition(pl *Player, s *Simulator) bool {
	onboarded := pl.StateBool("onboarded")
	return !onboarded
}
func (a OnboardAction) Exec(pl *Player, s *Simulator) error {
	pl.SetState("onboarded", true)
	return nil
}
func (a OnboardAction) Attributes(pl *Player, s *Simulator) map[string]any {
	attributes := map[string]any{
		"user.id": pl.id,
	}
	return attributes
}

type LoginAction struct{}

func (ac LoginAction) Name() string { return "login" }
func (ac LoginAction) Condition(pl *Player, s *Simulator) bool {
	if onboarded := pl.StateBool("onboarded"); onboarded {
		if loggedIn := pl.StateBool("loggedIn"); !loggedIn {
			return true
		}
	}
	return false
}
func (ac LoginAction) Exec(pl *Player, s *Simulator) error {
	pl.SetState("loggedIn", true)
	return nil
}
func (a LoginAction) Attributes(pl *Player, s *Simulator) map[string]any {
	attributes := map[string]any{
		"user.id": pl.id,
	}
	return attributes
}

type LogoutAction struct{}

func (ac LogoutAction) Name() string { return "logout" }
func (ac LogoutAction) Condition(pl *Player, s *Simulator) bool {
	if onboarded := pl.StateBool("onboarded"); onboarded {
		if loggedIn := pl.StateBool("loggedIn"); loggedIn {
			return true
		}
	}
	return false
}
func (ac LogoutAction) Exec(pl *Player, s *Simulator) error {
	pl.SetState("loggedIn", false)
	return nil
}
func (a LogoutAction) Attributes(pl *Player, s *Simulator) map[string]any {
	attributes := map[string]any{
		"user.id": pl.id,
	}
	return attributes
}
