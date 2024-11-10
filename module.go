package k9amqp

import (
	"go.k6.io/k6/js/common"
	"go.k6.io/k6/js/modules"
)

type (
	RootModule     struct{}
	ModuleInstance struct {
		vu     modules.VU
		k9amqp *K9amqp
	}
)

func init() {
	modules.Register("k6/x/k9amqp", New())
	modules.Register("k6/x/k9amqp/queue", new(Queue))
	modules.Register("k6/x/k9amqp/exchange", new(Exchange))
}

var (
	_ modules.Instance = &ModuleInstance{}
	_ modules.Module   = &RootModule{}
)

func New() *RootModule {
	return &RootModule{}
}

func (*RootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	metrics, err := registerMetrics(vu)
	if err != nil {
		common.Throw(vu.Runtime(), err)
	}
	return &ModuleInstance{
		vu:     vu,
		k9amqp: &K9amqp{vu: vu, metrics: metrics},
	}
}

func (mi *ModuleInstance) Exports() modules.Exports {
	return modules.Exports{
		Default: mi.k9amqp,
	}
}
