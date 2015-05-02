package rabbitmonit

import (
	"sort"

	"github.com/c-datculescu/rabbit-hole"
)

type lessFunc func(p1, p2 *rabbithole.QueueInfo) bool

type Ops struct {
	Host     string // the host to connect including the port
	Login    string // the username that allows us to retrieve statistics
	Password string // password for the username
}

/*
client returns a *rabbithole.Client on which we can run various operation types
*/
func (p *Ops) client() *rabbithole.Client {
	client, err := rabbithole.NewClient(p.Host, p.Login, p.Password)
	if err != nil {
		panic(err.Error())
	}

	return client
}

/*
Vhost is a small indirection which returns a vhost
*/
func (p *Ops) Vhost(vhost string) rabbithole.VhostInfo {
	client := p.client()
	vhostRet, err := client.GetVhost("/" + vhost)

	if err != nil {
		panic(err.Error())
	}

	return *vhostRet
}

/*
Nodes returns information about the current cluster individual nodes status
*/
func (p *Ops) Nodes() (returnNodes []NodeProperties) {
	client := p.client()

	nodes, err := client.ListNodes()
	if err != nil {
		panic(err.Error())
	}

	for _, node := range nodes {
		localNode := NodeProperties{
			NodeInfo: node,
		}

		localNode.Calculate()

		returnNodes = append(returnNodes, localNode)
	}

	return
}

/*
AccumulationQueues returns the top 10 most offending queues which can be a risk for the
cluster health
*/
func (p *Ops) AccumulationQueues() []QueueProperties {
	client := p.client()

	mapExtendedQueues := make([]QueueProperties, 0)

	queues, err := client.ListQueues()
	if err != nil {
		panic(err.Error())
	}

	qs := &queueSorter{}
	qs.Sort(queues)

	for _, queue := range queues {
		extQueue := new(QueueProperties)
		extQueue.Client = client
		extQueue.QueueInfo = queue
		extQueue.Calculate()

		mapExtendedQueues = append(mapExtendedQueues, *extQueue)
		if len(mapExtendedQueues) == 10 {
			break
		}
	}

	return mapExtendedQueues
}

/*
queue returns details about a queue from the api
*/
func (p *Ops) Queue(vhost, queue string) QueueProperties {
	client := p.client()
	queueDetail, err := client.GetQueue("/"+vhost, queue)
	if err != nil {
		panic(err.Error())
	}

	retQueue := &QueueProperties{
		QueueInfo: *queueDetail,
		Client:    client,
	}

	retQueue.Calculate()
	return *retQueue
}

/*
queues returns all the queues from a vhost along with detailed information about them
*/
func (p *Ops) Queues(vhost string) []QueueProperties {
	client := p.client()
	queues, err := client.ListQueuesIn("/" + vhost)
	if err != nil {
		panic(err.Error())
	}

	mapExtendedQueues := make([]QueueProperties, 0)

	for _, q := range queues {
		extQueue := new(QueueProperties)
		extQueue.QueueInfo = q
		extQueue.Client = client
		extQueue.Calculate()

		mapExtendedQueues = append(mapExtendedQueues, *extQueue)
	}

	return mapExtendedQueues
}

/*
queueSorter is used to sort a list of queues based on the number of messages accumulated
*/
type queueSorter struct {
	queues []rabbithole.QueueInfo
	less   []lessFunc
}

func (qs *queueSorter) Sort(queues []rabbithole.QueueInfo) {
	qs.queues = queues
	sort.Sort(qs)
}

func (qs *queueSorter) Len() int {
	return len(qs.queues)
}

func (qs *queueSorter) Swap(i, j int) {
	qs.queues[i], qs.queues[j] = qs.queues[j], qs.queues[i]
}

func (qs *queueSorter) Less(i, j int) bool {
	first, second := &qs.queues[i], &qs.queues[j]
	if first.MessagesRdy > second.MessagesRdy {
		return true
	}

	if first.MessagesRdy == second.MessagesRdy {
		return true
	}

	return false
}
