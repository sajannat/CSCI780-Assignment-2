package agent

import (
	"gitlab.com/akita/akita/v3/sim"
)

var EndDayHookPos = &sim.HookPos{Name: "End Day Hook"}

type EndDayHook struct {
	Cost int
}

func (h *EndDayHook) Func(hookCtx sim.HookCtx) {
	if hookCtx.Pos != EndDayHookPos {
		return
	}
	agent := hookCtx.Domain.(*Agent)
	dayCost := agent.Inventory * agent.InventoryCost

	if agent.IsRetailer {
		dayCost += agent.Backlog * agent.LostCustomerPenalty
		agent.Backlog = 0
	}

	h.Cost += dayCost
}
