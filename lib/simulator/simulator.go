package simulator

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

type Simulator struct {
	cfg     *Config
	players []*Player
	turn    int64
}

func NewSimulator(cfgPath string) (*Simulator, error) {
	cfg, err := parseConfig(cfgPath)
	if err != nil {
		return nil, err
	}

	s := &Simulator{
		cfg:  cfg,
		turn: 0,
	}

	for i := range cfg.Players.Count {
		s.players = append(s.players, NewPlayer(i, s))
	}

	return s, nil
}

// func (s *Simulator) Now() int64 { return s.turn * int64(s.cfg.TimeBetweenTurn) }

func (s *Simulator) NextTurn(ctx context.Context, out chan<- PlayerAction) error {
	if out == nil {
		return fmt.Errorf("channel can't be nil")
	}
	for range s.cfg.ActionPerTurn {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		pl, ac := s.findAction()

		if !s.playerTakesAction(pl, ac) {
			continue
		}

		if err := ac.Exec(pl, s); err != nil {
			return err
		}

		select {
		case out <- NewPlayerAction(pl, ac, s):
		default:
		}
	}
	s.turn++

	return nil
}

func (s *Simulator) NextNTurn(n int, out chan<- PlayerAction) error {
	for range n {
		if err := s.NextTurn(context.Background(), out); err != nil {
			return err
		}
		time.Sleep(time.Duration(s.cfg.TimeBetweenTurn) * time.Millisecond)
	}
	return nil
}

func (s *Simulator) Continuous(ctx context.Context, out chan<- PlayerAction, errCh chan<- error) {
	for {
		select {
		case <-ctx.Done():
			errCh <- ctx.Err()
			close(errCh)
			return
		default:
		}

		if err := s.NextTurn(ctx, out); err != nil {
			errCh <- err
		}

		time.Sleep(time.Duration(s.cfg.TimeBetweenTurn) * time.Millisecond)
	}
}

func (s *Simulator) findAction() (*Player, Action) {
	pl := s.players[rand.Intn(s.cfg.Players.Count)]
	ac := actionSlice[rand.Intn(len(actionSlice))]
	return pl, ac
}

func (s *Simulator) playerTakesAction(pl *Player, ac Action) bool {
	if rand.Intn(100) < s.weightForAction(ac) {
		return false
	}

	if !ac.Condition(pl, s) {
		return false
	}

	return true
}

func (s *Simulator) weightForAction(ac Action) int {
	acc, ok := s.cfg.Actions[ac.Name()]
	if !ok {
		return 100
	}
	acw := acc.Weight
	return acw
}
