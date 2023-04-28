package agent

import (
	"gitlab.com/akita/akita/v3/sim"
)

type Agent struct {
	*sim.TickingComponent

	Engine sim.Engine

	UpStream, DownStream   sim.Port
	UpStreamAgent, DownStreamAgent   sim.Port
	Inventory, Backlog, OnOrder   int
	InventoryCost, LostCustomerPenalty   int
	IsRetailer, IsFactory    bool
}

func NewAgent(name string, engine sim.Engine, inventory int) *Agent {
	a := &Agent{
		Inventory: inventory,
	}
	a.TickingComponent = sim.NewTickingComponent(name, engine, 1*sim.Hz, a)
	a.UpStream = sim.NewLimitNumMsgPort(a, 4, name+".UpStream")
	a.DownStream = sim.NewLimitNumMsgPort(a, 4, name+".DownStream")
	return a
}

func (a *Agent) Tick(now sim.VTimeInSec) bool {
	a.receiveOrder(now)
	a.receiveShipment(now)
	a.sendShipment(now)
	a.sendOrder(now)

	hookCtx := sim.HookCtx{Pos: EndDayHookPos, Domain: a}
	a.InvokeHook(hookCtx)

	return false
}

func (a *Agent) receiveOrder(now sim.VTimeInSec) {
	msg := a.DownStream.Retrieve(now)
	if msg == nil {
		return
	}
	order, ok := msg.(*Order)
	if !ok {
		return
	}
	a.Backlog += order.Quantity
	println(a.Name(), "receives order ", order.Quantity)
}

func (a *Agent) receiveShipment(now sim.VTimeInSec) {
	msg := a.UpStream.Retrieve(now)
	if msg == nil {
		return
	}
	shipment, ok := msg.(*Shipment)
	if !ok {
		return
	}
	a.Inventory += shipment.Quantity
	println(a.Name(), "receives shipment ", shipment.Quantity)
}

func (a *Agent) sendShipment(now sim.VTimeInSec) {
	if a.Inventory == 0 || a.Backlog == 0 {
		return
	}

	shipmentQuantity := min(a.Inventory, a.Backlog)
	a.Inventory -= shipmentQuantity
	a.Backlog -= shipmentQuantity

	if !a.IsRetailer {
		shipment := &Shipment{
			Quantity: shipmentQuantity,
			MsgMeta: sim.MsgMeta{
				Src:      a.DownStream,
				Dst:      a.DownStreamAgent,
				SendTime: now,
			},
		}
		println(a.Name(), "sends shipment ", shipmentQuantity)
		a.DownStream.Send(shipment)
	}
}

func (a *Agent) sendOrder(now sim.VTimeInSec) {
	orderQuantity := a.Backlog - a.Inventory - a.OnOrder
	if orderQuantity <= 0 {
		return
	}

	order := &Order{
		Quantity: orderQuantity,
		MsgMeta: sim.MsgMeta{
			Src:      a.UpStream,
			Dst:      a.UpStreamAgent,
			SendTime: now,
		},
	}

	a.OnOrder += orderQuantity

	if !a.IsFactory {
		println(a.Name(), "sends order ", orderQuantity)
		a.UpStream.Send(order)
	} else {
		a.Inventory += orderQuantity
	}
}

func SetAgentCosts(a *Agent, inventoryCost, lostCustomerPenalty int) {
	a.InventoryCost = inventoryCost
	a.LostCustomerPenalty = lostCustomerPenalty
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
