package rabbitmonit

import (
	"github.com/c-datculescu/rabbit-hole"
)

/*
VhostProperties extends rabbithole.VhostInfo with additional alerting/status values
*/
type VhostProperties struct {
	VhostInfo rabbithole.VhostInfo
	Error     VhostAlert
	Warning   VhostAlert
	Stats     VhostStats
}

/*
VhostStats represents a collectional of additional stats that are calculated for the given vhost
*/
type VhostStats struct {
	EnqueueDequeueDiff float32
}

/*
VhostAlert contains all the alerts that a vhost can report
*/
type VhostAlert struct {
	Rdy            bool // are there ready messages available
	Has            bool // do we have any errors
	ConsumptionLow bool // the enqueue rate is bigger than the dequeue rate
}

/*
Calculate runs all the statistics on the currently given vhost
*/
func (vp *VhostProperties) Calculate() {
	vp.Error = VhostAlert{}
	vp.Warning = VhostAlert{}
	vp.Stats = VhostStats{}

	vp.alertRdy().
		alertConsumptionLow()
}

/*
alertRdy raises an alert/warning when messages ready exceed a certain limit

threshold for alert is 1000

threshold for warning is 0
*/
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

/*
alertConsumptionLow raises an alert/warning when consumption rate vs ingestion rate exceeds certain values

threshold for alert is rate difference bigger than 10 and consuption is less than publishing

threshold for alert is rate difference bigger than 5 and consuption is less than publishing
*/
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
