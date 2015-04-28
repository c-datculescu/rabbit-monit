package rabbitmonit

import (
	"sort"
	"strconv"

	"github.com/michaelklishin/rabbit-hole"
)

type lessFunc func(p1, p2 *rabbithole.QueueInfo) bool

type Ops struct {
	Host     string // the host to connect including the port
	Login    string // the username that allows us to retrieve statistics
	Password string // password for the username
}

type QueueSorter struct {
	queues []rabbithole.QueueInfo
	less   []lessFunc
}

func (qs *QueueSorter) Sort(queues []rabbithole.QueueInfo) {
	qs.queues = queues
	sort.Sort(qs)
}

func (qs *QueueSorter) Len() int {
	return len(qs.queues)
}

func (qs *QueueSorter) Swap(i, j int) {
	qs.queues[i], qs.queues[j] = qs.queues[j], qs.queues[i]
}

func (qs *QueueSorter) Less(i, j int) bool {
	first, second := &qs.queues[i], &qs.queues[j]
	if first.MessagesReady > second.MessagesReady {
		return true
	}

	if first.MessagesReady == second.MessagesReady {
		return true
	}

	return false
}

type NodeProperties struct {
	NodeInfo          rabbithole.NodeInfo
	FdPercentage      float32 // what percentage from the file descriptors are used
	DiskPercentage    float64 // that percentage of disk is used on the database partition
	MemPercentage     float64 // what percentage of memory is used
	ErlProcPercentage float64 // that percentage of erlang processes are used
	SockPercentage    float64 // what percentage of cpu is used
}

type QueueProperties struct {
	QueueInfo               rabbithole.QueueInfo
	AlertState              bool
	AlertNonDurable         bool
	AlertRdy                bool
	AlertUnack              bool
	AlertListener           bool
	AlertUtilization        bool
	AlertIntake             bool
	AlertNonDurableMessages bool
	NonPersistentMessages   int
}

func (qp *QueueProperties) Calculate() {
	if qp.QueueInfo.State != "running" {
		qp.AlertState = true
	}

	if qp.QueueInfo.MessagesRam > 0 {
		qp.AlertNonDurable = true
	}

	if qp.QueueInfo.MessagesReady > 100 {
		qp.AlertRdy = true
	}

	if qp.QueueInfo.MessagesUnacknowledged > 100 {
		qp.AlertUnack = true
	}

	if qp.QueueInfo.Consumers == 0 {
		qp.AlertListener = true
	}

	if qp.QueueInfo.ConsumerUtilisation == "" {
		qp.QueueInfo.ConsumerUtilisation = 0
	}

	var consUtil float64
	switch qp.QueueInfo.ConsumerUtilisation.(type) {
	case string:
		consUtil, _ = strconv.ParseFloat(qp.QueueInfo.ConsumerUtilisation.(string), 64)
	case int:
		consUtil = float64(qp.QueueInfo.ConsumerUtilisation.(int))
	default:
		consUtil = qp.QueueInfo.ConsumerUtilisation.(float64)
	}

	if consUtil < 70 {
		qp.AlertUtilization = true
	}

	if qp.QueueInfo.MessagesReadyDetails.Rate > 1 {
		qp.AlertIntake = true
	}

	if qp.QueueInfo.MessagesRam-qp.QueueInfo.MessagesPersitent > 0 {
		qp.AlertNonDurableMessages = true
		qp.NonPersistentMessages = qp.QueueInfo.MessagesRam - qp.QueueInfo.MessagesPersitent
	}
}

// retrieves a client on which we can run some operations
func (p *Ops) client() *rabbithole.Client {
	client, err := rabbithole.NewClient(p.Host, p.Login, p.Password)
	if err != nil {
		panic(err.Error())
	}

	return client
}

/*
getClusterNodes returns slightly better information about the monitored nodes in the cluster
with the intent of being able to plot the information
*/
func (p *Ops) getClusterNodes() (returnNodes []*NodeProperties) {
	client := p.client()

	nodes, err := client.ListNodes()
	if err != nil {
		panic(err.Error())
	}

	for _, node := range nodes {
		localNode := &NodeProperties{
			NodeInfo: node,
		}

		res := (float64(node.FdUsed) / float64(node.FdTotal)) * 100
		localNode.FdPercentage = float32(roundPlus(res, 2))

		res = (float64(node.DiskFreeLimit) / float64(node.DiskFree)) * 100
		localNode.DiskPercentage = roundPlus(float64(100-float32(res)), 2)

		res = (float64(node.MemUsed) / float64(node.MemLimit)) * 100
		localNode.MemPercentage = roundPlus(float64(res), 2)

		res = (float64(node.ProcUsed) / float64(node.ProcTotal)) * 100
		localNode.ErlProcPercentage = roundPlus(float64(res), 2)

		res = (float64(node.SocketsUsed) / float64(node.SocketsTotal)) * 100
		localNode.SockPercentage = roundPlus(float64(res), 2)

		returnNodes = append(returnNodes, localNode)
	}

	return
}

func (p *Ops) getAccumulationQueues() []QueueProperties {
	client := p.client()

	mapExtendedQueues := make([]QueueProperties, 0)

	queues, err := client.ListQueues()
	if err != nil {
		panic(err.Error())
	}

	qs := &QueueSorter{}
	qs.Sort(queues)

	for _, queue := range queues {
		extQueue := new(QueueProperties)
		extQueue.QueueInfo = queue
		extQueue.Calculate()

		mapExtendedQueues = append(mapExtendedQueues, *extQueue)
	}

	return mapExtendedQueues
}
