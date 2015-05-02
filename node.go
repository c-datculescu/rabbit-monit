package rabbitmonit

import (
	"github.com/c-datculescu/rabbit-hole"
)

/*
nodeProperties is a structure offering slightly more flexibility/statistics than the rabbit-hole struct
*/
type NodeProperties struct {
	Stats    NodeStat
	Error    NodeAlert
	Warning  NodeAlert
	NodeInfo rabbithole.NodeInfo
}

/*
NodeStat Holds all the relevant node statistics.
@todo add more statistics in the future as well as warnings
*/
type NodeStat struct {
	FdUsedPercentage   float64 // percentage of used file descriptors
	DiskUsedPercentage float64 // percentage of disk used until alarm limit
	MemUsedPercentage  float64 // percentage of memory used
	ErlUsedPercentage  float64 // percentage of erlang processes used
	SockUsedpercentage float64 // percentage of sockets used
}

/*
NodeAlert holds all the needed alerts for critical issues related to the current node
@todo identify other alerts which might be relevant
*/
type NodeAlert struct {
	Fd     bool // file descriptors. > 80 = warning, > 90 = error
	Erl    bool // erlang processes. > 80 = warning, > 90 = error
	Mem    bool // memory. > 85 = warning, > 95 = error
	Hdd    bool // disk space. > 80 = warning, > 90 = error
	Sock   bool // sockets used. > 80 = warning, > 90 = error
	Status bool // status of the node. if status is not "running", error
}

func (np *NodeProperties) statsFd() *NodeProperties {
	res := (float64(np.NodeInfo.FdUsed) / float64(np.NodeInfo.FdTotal)) * 100
	np.Stats.FdUsedPercentage = float64(roundPlus(res, 2))
	return np
}

func (np *NodeProperties) statsDisk() *NodeProperties {
	res := (float64(np.NodeInfo.DiskFreeLimit) / float64(np.NodeInfo.DiskFree)) * 100
	np.Stats.DiskUsedPercentage = float64(roundPlus(res, 2))
	return np
}

func (np *NodeProperties) statsMem() *NodeProperties {
	res := (float64(np.NodeInfo.MemUsed) / float64(np.NodeInfo.MemLimit)) * 100
	np.Stats.MemUsedPercentage = float64(roundPlus(res, 2))
	return np
}

func (np *NodeProperties) statsErl() *NodeProperties {
	res := (float64(np.NodeInfo.ProcUsed) / float64(np.NodeInfo.ProcTotal)) * 100
	np.Stats.ErlUsedPercentage = float64(roundPlus(res, 2))
	return np
}

func (np *NodeProperties) statsSock() *NodeProperties {
	res := (float64(np.NodeInfo.SocketsUsed) / float64(np.NodeInfo.SocketsTotal)) * 100
	np.Stats.SockUsedpercentage = roundPlus(float64(res), 2)
	return np
}

func (np *NodeProperties) alertFd() *NodeProperties {
	if np.Stats.FdUsedPercentage > 90 {
		np.Error.Fd = true
	} else if np.Stats.FdUsedPercentage > 80 {
		np.Warning.Fd = true
	}
	return np
}

func (np *NodeProperties) alertErl() *NodeProperties {
	if np.Stats.ErlUsedPercentage > 90 {
		np.Error.Erl = true
	} else if np.Stats.ErlUsedPercentage > 80 {
		np.Warning.Erl = true
	}
	return np
}

func (np *NodeProperties) alertMem() *NodeProperties {
	if np.Stats.MemUsedPercentage > 90 {
		np.Error.Mem = true
	} else if np.Stats.MemUsedPercentage > 85 {
		np.Warning.Mem = true
	}

	return np
}

func (np *NodeProperties) alertHdd() *NodeProperties {
	if np.Stats.DiskUsedPercentage > 90 {
		np.Error.Hdd = true
	} else if np.Stats.DiskUsedPercentage > 80 {
		np.Warning.Hdd = true
	}
	return np
}

func (np *NodeProperties) alertSock() *NodeProperties {
	if np.Stats.SockUsedpercentage > 90 {
		np.Error.Sock = true
	} else if np.Stats.SockUsedpercentage > 80 {
		np.Warning.Sock = true
	}
	return np
}

func (np *NodeProperties) alertStatus() *NodeProperties {
	if np.NodeInfo.Running == false {
		np.Error.Status = true
	}
	return np
}

func (np *NodeProperties) Calculate() {
	np.statsFd().
		statsDisk().
		statsMem().
		statsErl().
		statsSock().
		alertFd().
		alertErl().
		alertMem().
		alertHdd().
		alertSock().
		alertStatus()
}
