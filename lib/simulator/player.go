package simulator

import (
	"maps"
	"math/rand"
)

// A Player is a participant in the Simulator, which can be thought of as an RPG where Players perform Actions each turn.
type Player struct {
	id    int
	state map[string]any
}

func (pl *Player) State(key string) any {
	v, ok := pl.state[key]
	if !ok {
		return nil
	}
	return v
}

func (pl *Player) StateBool(key string) bool {
	v, ok := pl.state[key].(bool)
	if !ok {
		return false
	}
	return v
}

func (pl *Player) StateString(key string) string {
	v, ok := pl.state[key].(string)
	if !ok {
		return ""
	}
	return v
}

func (pl *Player) StateInt(key string) int {
	v, ok := pl.state[key].(int)
	if !ok {
		return 0
	}
	return v
}

func (pl *Player) SetState(key string, value any) {
	pl.state[key] = value
}

func NewPlayer(id int, s *Simulator) *Player {
	pl := Player{
		id:    id,
		state: makeInitState(id, s),
	}

	return &pl
}

func makeInitState(id int, s *Simulator) map[string]any {
	state := map[string]any{}
	initStates := s.cfg.Players.InitStates
	if len(initStates) > 0 {
		total := 0
		for _, sc := range initStates {
			total += sc.Weight
		}
		r := rand.Intn(total)
		acc := 0
		chosen := initStates[0]
		for _, sc := range initStates {
			acc += sc.Weight
			if r < acc {
				chosen = sc
				break
			}
		}
		maps.Copy(state, chosen.State)
	}
	state["user.id"] = id
	return state
}
