package agent

import (
	"gitlab.com/akita/akita/v3/sim"
)

type Order struct {
	sim.MsgMeta
	Quantity int
}

func (o *Order) Meta() *sim.MsgMeta {
	return &o.MsgMeta
}

type Shipment struct {
	sim.MsgMeta
	Quantity int
}

func (s *Shipment) Meta() *sim.MsgMeta {
	return &s.MsgMeta
}
