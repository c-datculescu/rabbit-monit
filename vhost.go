package rabbitmonit

import (
	"github.com/c-datculescu/rabbit-hole"
)

type VhostProperties struct {
	VhostInfo rabbithole.VhostInfo
	Error     VhostAlert
	Warning   VhostAlert
	Stats     VhostStats
}

type VhostStats struct {
	EnqueueDequeueDiff float32
}

type VhostAlert struct {
	Rdy            bool // are there ready messages available
	Has            bool // do we have any errors
	ConsumptionLow bool // the enqueue rate is bigger than the dequeue rate
}

func (vp *VhostProperties) Calculate() {
	vp.Error = VhostAlert{}
	vp.Warning = VhostAlert{}
	vp.Stats = VhostStats{}

	vp.alertRdy().
		alertConsumptionLow()
}

func (vp *VhostProperties) alertRdy() *VhostProperties {
	if vp.VhostInfo.MessagesRdy > 1000 {
		vp.Error.Has = true
		vp.Error.Rdy = true
	} else if vp.VhostInfo.MessagesRdy > 0 {
		vp.Warning.Has = true
		vp.Warning.Rdy = true
	}

	return vp
}

func (vp *VhostProperties) alertConsumptionLow() *VhostProperties {
	rate := vp.VhostInfo.MessageStats.PublishDetails.Rate - vp.VhostInfo.MessageStats.DeliverDetails.Rate
	vp.Stats.EnqueueDequeueDiff = rate
	lowerThan := vp.VhostInfo.MessageStats.DeliverDetails.Rate < vp.VhostInfo.MessageStats.PublishDetails.Rate
	if lowerThan && rate > 10 {
		vp.Error.Has = true
		vp.Error.ConsumptionLow = true
	} else if lowerThan && rate > 5 {
		vp.Warning.Has = true
		vp.Warning.ConsumptionLow = true
	}

	return vp
}
