package rabbitmonit

import (
	"sort"
	"strconv"

	"github.com/c-datculescu/rabbit-hole"
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
	if first.MessagesRdy > second.MessagesRdy {
		return true
	}

	if first.MessagesRdy == second.MessagesRdy {
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
	AlertFd           int
	AlertErl          int
	AlertMem          int
	AlertHdd          int
	AlertSock         int
	AlertStatus       int
}

type QueueProperties struct {
	QueueInfo               rabbithole.QueueInfo
	AlertState              int
	AlertNonDurable         int
	AlertRdy                int
	AlertUnack              int
	AlertListener           int
	AlertUtilization        int
	AlertIntake             int
	AlertNonDurableMessages int
	NonPersistentMessages   int
}

func (qp *QueueProperties) Calculate() {
	qp.AlertState = 0
	if qp.QueueInfo.State != "running" {
		qp.AlertState = 1
	}

	qp.AlertNonDurable = 0
	if qp.QueueInfo.Durable == false {
		qp.AlertNonDurable = 1
	}

	qp.AlertRdy = 0
	if qp.QueueInfo.MessagesRdy > 100 {
		qp.AlertRdy = 1
	} else if qp.QueueInfo.MessagesRdy > 0 {
		qp.AlertRdy = 2
	}

	qp.AlertUnack = 0

	qp.AlertListener = 0
	if qp.QueueInfo.Consumers == 0 && qp.QueueInfo.MessagesRdy > 0 {
		qp.AlertListener = 1
	} else if qp.QueueInfo.Consumers < 3 && qp.QueueInfo.MessagesRdy > 0 {
		qp.AlertListener = 2
	}

	var consUtil float64
	switch qp.QueueInfo.ConsumerUtilisation.(type) {
	case string:
		consUtil, _ = strconv.ParseFloat(qp.QueueInfo.ConsumerUtilisation.(string), 64)
		if qp.QueueInfo.ConsumerUtilisation == "" {
			qp.QueueInfo.ConsumerUtilisation = "0"
		}
	case int:
		consUtil = float64(qp.QueueInfo.ConsumerUtilisation.(int))
	case float64:
		consUtil = qp.QueueInfo.ConsumerUtilisation.(float64)
	}

	qp.AlertUtilization = 0
	if consUtil < 30 && qp.QueueInfo.MessagesRdy > 0 {
		qp.AlertUtilization = 1
	} else if consUtil < 70 && qp.QueueInfo.MessagesRdy > 0 {
		qp.AlertUtilization = 2
	}

	qp.AlertIntake = 0
	if qp.QueueInfo.MessagesRdyDetails.Rate > 1 {
		qp.AlertIntake = 2
	}

	qp.AlertNonDurableMessages = 0
	if qp.QueueInfo.MessagesRam-qp.QueueInfo.MessagesPersistent > 0 {
		qp.AlertNonDurableMessages = 1
		qp.NonPersistentMessages = qp.QueueInfo.MessagesRam - qp.QueueInfo.MessagesPersistent
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

func (p *Ops) Vhost(vhost string) rabbithole.VhostInfo {
	client := p.client()
	vhostRet, err := client.GetVhost("/" + vhost)

	if err != nil {
		panic(err.Error())
	}

	return *vhostRet
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
		localNode.DiskPercentage = roundPlus(float64(res), 2)

		res = (float64(node.MemUsed) / float64(node.MemLimit)) * 100
		localNode.MemPercentage = roundPlus(float64(res), 2)

		res = (float64(node.ProcUsed) / float64(node.ProcTotal)) * 100
		localNode.ErlProcPercentage = roundPlus(float64(res), 2)

		res = (float64(node.SocketsUsed) / float64(node.SocketsTotal)) * 100
		localNode.SockPercentage = roundPlus(float64(res), 2)

		// calculate alerts
		localNode.AlertFd = 0
		if localNode.FdPercentage > 90 {
			localNode.AlertFd = 1
		} else if localNode.FdPercentage > 80 {
			localNode.AlertFd = 2
		}

		localNode.AlertErl = 0
		if localNode.ErlProcPercentage > 90 {
			localNode.AlertErl = 1
		} else if localNode.ErlProcPercentage > 80 {
			localNode.AlertErl = 2
		}

		localNode.AlertMem = 0
		if localNode.MemPercentage > 90 {
			localNode.AlertMem = 1
		} else if localNode.MemPercentage > 85 {
			localNode.AlertMem = 2
		}

		localNode.AlertHdd = 0
		if localNode.DiskPercentage > 90 {
			localNode.AlertHdd = 1
		} else if localNode.DiskPercentage > 80 {
			localNode.AlertHdd = 2
		}

		localNode.AlertSock = 0
		if localNode.SockPercentage > 90 {
			localNode.AlertSock = 1
		} else if localNode.SockPercentage > 80 {
			localNode.AlertSock = 2
		}

		localNode.AlertStatus = 0
		if localNode.NodeInfo.Running == false {
			localNode.AlertStatus = 1
		}

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
		if len(mapExtendedQueues) == 10 {
			break
		}
	}

	return mapExtendedQueues
}

func (p *Ops) getQueue(vhost, queue string) QueueProperties {
	client := p.client()
	queueDetail, err := client.GetQueue("/"+vhost, queue)
	if err != nil {
		panic(err.Error())
	}

	retQueue := &QueueProperties{
		QueueInfo: *queueDetail,
	}

	retQueue.Calculate()
	return *retQueue
}

func (p *Ops) getQueues(vhost string) []QueueProperties {
	client := p.client()
	queues, err := client.ListQueuesIn("/" + vhost)
	if err != nil {
		panic(err.Error())
	}

	mapExtendedQueues := make([]QueueProperties, 0)

	for _, q := range queues {
		extQueue := new(QueueProperties)
		extQueue.QueueInfo = q
		extQueue.Calculate()

		mapExtendedQueues = append(mapExtendedQueues, *extQueue)
	}

	return mapExtendedQueues
}
