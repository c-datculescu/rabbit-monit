package rabbitmonit

import (
	"github.com/c-datculescu/rabbit-hole"
)

type VhostProperties struct {
	VhostInfo rabbithole.VhostInfo
	Error     VhostAlert
	Warning   VhostAlert
}

type VhostAlert struct {
	Rdy bool // are there ready messages available
	Has bool // do we have any errors
}

func (vp *VhostProperties) Calculate() {
	vp.Error = VhostAlert{}
	vp.Warning = VhostAlert{}

	vp.alertRdy()
}

func (vp *VhostProperties) alertRdy() {
	if vp.VhostInfo.MessagesRdy > 1000 {
		vp.Error.Has = true
		vp.Error.Rdy = true
	} else if vp.VhostInfo.MessagesRdy > 0 {
		vp.Warning.Has = true
		vp.Warning.Rdy = true
	}
}
