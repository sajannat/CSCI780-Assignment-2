package beergame

import (
	"beer-game/agent"

	"gitlab.com/akita/akita/v3/sim"
)

type NewCustomerEvent struct {
	time     sim.VTimeInSec
	handler  sim.Handler
	quantity int
}

func (e *NewCustomerEvent) Handler() sim.Handler {
	return e.handler
}

func (e *NewCustomerEvent) Time() sim.VTimeInSec {
	return e.time
}

func (e *NewCustomerEvent) IsSecondary() bool {
	return false
}

type NewCustomerEventHandler struct {
	Retailer *agent.Agent
}

func (h *NewCustomerEventHandler) Handle(e sim.Event) error {
	order := &agent.Order{
		Quantity: e.(*NewCustomerEvent).quantity,
	}
	h.Retailer.DownStream.Recv(order)
	return nil
}

func Beer() {
	engine := sim.NewSerialEngine()

	factory := agent.NewAgent("Factory", engine, 10)
	distributor := agent.NewAgent("Distributor", engine, 10)
	wholesaler := agent.NewAgent("Wholesaler", engine, 10)
	retailer := agent.NewAgent("Retailer", engine, 10)

	agent.SetAgentCosts(factory, 1, 0)     // Factory does not have lost customer penalty
	agent.SetAgentCosts(distributor, 2, 0) // Distributor does not have lost customer penalty
	agent.SetAgentCosts(wholesaler, 3, 0)  // Wholesaler does not have lost customer penalty
	agent.SetAgentCosts(retailer, 4, 30)   // Retailer has lost customer penalty

	conn := sim.NewDirectConnection("Conn", engine, 1*sim.GHz)

	factory.IsFactory = true
	retailer.IsRetailer = true

	retailer.UpStreamAgent = wholesaler.DownStream
	wholesaler.UpStreamAgent = distributor.DownStream
	distributor.UpStreamAgent = factory.DownStream

	wholesaler.DownStreamAgent = retailer.UpStream
	distributor.DownStreamAgent = wholesaler.UpStream
	factory.DownStreamAgent = distributor.UpStream

	conn.PlugIn(factory.UpStream, 1)
	conn.PlugIn(distributor.UpStream, 1)
	conn.PlugIn(wholesaler.UpStream, 1)
	conn.PlugIn(retailer.UpStream, 1)

	conn.PlugIn(factory.DownStream, 1)
	conn.PlugIn(distributor.DownStream, 1)
	conn.PlugIn(wholesaler.DownStream, 1)
	conn.PlugIn(retailer.DownStream, 1)

	eventHandler := &NewCustomerEventHandler{
		Retailer: retailer,
	}

	for i := 0; i < 100; i++ {
		cycle := sim.VTimeInSec(i)
		quantity := 16
		if i < 8 {
			quantity = 4
		}

		event := &NewCustomerEvent{
			time:     cycle,
			handler:  eventHandler,
			quantity: quantity,
		}
		engine.Schedule(event)
	}

	costCounter := &agent.EndDayHook{Cost: 0}

	retailer.AcceptHook(costCounter)
	wholesaler.AcceptHook(costCounter)
	distributor.AcceptHook(costCounter)
	factory.AcceptHook(costCounter)

	engine.Run()

	println("Factory Inventory:", factory.Inventory)
	println("Distributor Inventory:", distributor.Inventory)
	println("Wholesaler Inventory:", wholesaler.Inventory)
	println("Retailer Inventory:", retailer.Inventory)

	println("Total Cost:", costCounter.Cost)
}
