# nux

nux, standing for Enhanced User Experience, is a solution for enriching user profiles based on rules applied to telemetry data.


## nux-cli

#### github.com/jimtang2/nux/cmd

- `nux init`
- `nux col`
- `nux simulator`
- `nux gen`

## Simulator

#### github.com/jimtang2/nux/lib/simulator

- defines a custom Discrete Event Simulator that runs by turn or continuously mode
- defines configuration of simulation size, participant states, actions, weights
- provides an Event channel to export to Kafka, OtelCollector, etc

#### github.com/jimtang2/nux/lib/simulator-actions

- defines implementations of simulator.Action interface
- package cex defines general actions for a CEX

#### github.com/jimtang2/nux/lib/otel/util

- defines an otel sender to send Simulator output Events to an otel collector

##### Abstractions

- Simulator 
- Config
- Player
- Action
- Event

## Infrastructure

- Otel Collector
- Kafka
- PostgreSQL