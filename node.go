package rabbitmonit

import "github.com/c-datculescu/rabbit-hole"

/*
NodeProperties is a structure offering slightly more flexibility/statistics than the rabbit-hole struct
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
	SockUsedPercentage float64 // percentage of sockets used
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

/*
statsFd calculates the stats related to file descriptors available for rabbitmq
*/
func (np *NodeProperties) statsFd() *NodeProperties {
	res := (float64(np.NodeInfo.FdUsed) / float64(np.NodeInfo.FdTotal)) * 100
	np.Stats.FdUsedPercentage = float64(RoundPlus(res, 2))
	return np
}

/*
statsDisk calculates stats regarding the disk space available for rabbitmq before raising the alert
*/
func (np *NodeProperties) statsDisk() *NodeProperties {
	res := (float64(np.NodeInfo.DiskFreeLimit) / float64(np.NodeInfo.DiskFree)) * 100
	np.Stats.DiskUsedPercentage = float64(RoundPlus(res, 2))
	return np
}

/*
statsMem calculates the stats of a rabbitmq node in regards to the limit before raising an alert
*/
func (np *NodeProperties) statsMem() *NodeProperties {
	res := (float64(np.NodeInfo.MemUsed) / float64(np.NodeInfo.MemLimit)) * 100
	np.Stats.MemUsedPercentage = float64(RoundPlus(res, 2))
	return np
}

/*
statsErl calculates the stats related to erlang processes available for rabbitmq before exhaustion
*/
func (np *NodeProperties) statsErl() *NodeProperties {
	res := (float64(np.NodeInfo.ProcUsed) / float64(np.NodeInfo.ProcTotal)) * 100
	np.Stats.ErlUsedPercentage = float64(RoundPlus(res, 2))
	return np
}

/*
statsSock calculates the stats related to sockets available for rabbitmq before exhaustion
*/
func (np *NodeProperties) statsSock() *NodeProperties {
	res := (float64(np.NodeInfo.SocketsUsed) / float64(np.NodeInfo.SocketsTotal)) * 100
	np.Stats.SockUsedPercentage = RoundPlus(float64(res), 2)
	return np
}

/*
alertFd calculates whether it should raise an alert or a warning for file descriptors

if file descriptors are over 90% than an alert is raised

if file descriptors are over 80% a warning gets raised
*/
func (np *NodeProperties) alertFd() *NodeProperties {
	if np.Stats.FdUsedPercentage > 90 {
		np.Error.Fd = true
	} else if np.Stats.FdUsedPercentage > 80 {
		np.Warning.Fd = true
	}
	return np
}

/*
alertErl caulculates whetger it should raise an alert or a warning for erlang processes availability

if erlang processes are over 90% it raises an alert

if erlang processes are over 80% it raises a warning
*/
func (np *NodeProperties) alertErl() *NodeProperties {
	if np.Stats.ErlUsedPercentage > 90 {
		np.Error.Erl = true
	} else if np.Stats.ErlUsedPercentage > 80 {
		np.Warning.Erl = true
	}
	return np
}

/*
alertMem calculates whether it should raise an alert or a warning for memory approaching the alert threshold

if memory is over 90% an alert is raised

if memory is over 85% a warning is raised
*/
func (np *NodeProperties) alertMem() *NodeProperties {
	if np.Stats.MemUsedPercentage > 90 {
		np.Error.Mem = true
	} else if np.Stats.MemUsedPercentage > 85 {
		np.Warning.Mem = true
	}

	return np
}

/*
alertHdd calculates whether there should be an alert or a warning for disk space approaching the alerting threshold

if disk space is over 90% an alert is raised

if disk space is over 80% an warning is raised
*/
func (np *NodeProperties) alertHdd() *NodeProperties {
	if np.Stats.DiskUsedPercentage > 90 {
		np.Error.Hdd = true
	} else if np.Stats.DiskUsedPercentage > 80 {
		np.Warning.Hdd = true
	}
	return np
}

/*
alertSock calculates whetjer there should be an alert or a warning for socket exhaustion

if socket consumption is 90% or over an alert is raised

if socket consumption is 80% or over a warning is raised
*/
func (np *NodeProperties) alertSock() *NodeProperties {
	if np.Stats.SockUsedPercentage > 90 {
		np.Error.Sock = true
	} else if np.Stats.SockUsedPercentage > 80 {
		np.Warning.Sock = true
	}
	return np
}

/*
alertStatus calculates whether there should be an alert for status of the node

if the node status is other than running, an alert will be raised
*/
func (np *NodeProperties) alertStatus() *NodeProperties {
	if np.NodeInfo.Running == false {
		np.Error.Status = true
	}
	return np
}

/*
Calculate performs various calculations and alerts discovery on top of the current node
and also runs all the stats/alert calculation functions
*/
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
