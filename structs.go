package rabbitmonit

import (
	"sort"

	"github.com/c-datculescu/rabbit-hole"
)

type lessFunc func(p1, p2 *rabbithole.QueueInfo) bool

/*
Ops is the main structure for monitoring operations over a rabbitmq cluster
*/
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
Vhosts returns a list of all vhosts in the current cluster and sorts them by warnings/errors
*/
func (p *Ops) Vhosts() []VhostProperties {
	client := p.client()
	vhostsRet, err := client.ListVhosts()

	if err != nil {
		panic(err.Error())
	}

	var mapVhosts []VhostProperties

	for _, vhost := range vhostsRet {
		vh := &VhostProperties{
			VhostInfo: vhost,
		}
		vh.Calculate()
		mapVhosts = append(mapVhosts, *vh)
	}

	vs := &vhostSorter{}
	vs.Sort(mapVhosts)

	return mapVhosts
}

/*
Vhost is a small indirection which returns a vhost
*/
func (p *Ops) Vhost(vhost string) rabbithole.VhostInfo {
	client := p.client()
	vhostRet, err := client.GetVhost(vhost)

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

	var mapExtendedQueues []QueueProperties

	queues, err := client.ListQueues()
	if err != nil {
		panic(err.Error())
	}

	for _, queue := range queues {
		extQueue := new(QueueProperties)
		extQueue.Client = client
		extQueue.QueueInfo = queue
		extQueue.Calculate()

		mapExtendedQueues = append(mapExtendedQueues, *extQueue)
	}

	qs := &queueSorter{}
	qs.Sort(mapExtendedQueues)

	return mapExtendedQueues
}

/*
Queue returns details about a queue from the api
*/
func (p *Ops) Queue(vhost, queue string) QueueProperties {
	client := p.client()
	queueDetail, err := client.GetQueue(vhost, queue)
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
Queues returns all the queues from a vhost along with detailed information about them
*/
func (p *Ops) Queues(vhost string) []QueueProperties {
	client := p.client()
	queues, err := client.ListQueuesIn(vhost)
	if err != nil {
		panic(err.Error())
	}

	var mapExtendedQueues []QueueProperties

	for _, q := range queues {
		extQueue := new(QueueProperties)
		extQueue.QueueInfo = q
		extQueue.Client = client
		extQueue.Calculate()

		mapExtendedQueues = append(mapExtendedQueues, *extQueue)
	}

	qs := &queueSorter{}
	qs.Sort(mapExtendedQueues)

	return mapExtendedQueues
}

/*
vhostSorter is responsible for sorting vhosts based on the warnings/errors they contain
*/
type vhostSorter struct {
	vhosts []VhostProperties
	less   []lessFunc
}

func (vs *vhostSorter) Sort(vhosts []VhostProperties) {
	vs.vhosts = vhosts
	sort.Sort(vs)
}

func (vs *vhostSorter) Len() int {
	return len(vs.vhosts)
}

func (vs *vhostSorter) Swap(i, j int) {
	vs.vhosts[i], vs.vhosts[j] = vs.vhosts[j], vs.vhosts[i]
}

func (vs *vhostSorter) Less(i, j int) bool {
	first, second := &vs.vhosts[i], &vs.vhosts[j]

	var firstValue, secondValue = 1, 1

	if first.Error.Has {
		firstValue = 4
	} else if first.Warning.Has {
		firstValue = 2
	}

	if second.Error.Has {
		secondValue = 4
	} else if second.Warning.Has {
		secondValue = 2
	}

	if firstValue > secondValue || first.VhostInfo.MessagesRdy > second.VhostInfo.MessagesRdy {
		return true
	}

	return false
}

/*
queueSorter is used to sort a list of queues based on the number of messages accumulated
*/
type queueSorter struct {
	queues []QueueProperties
	less   []lessFunc
}

func (qs *queueSorter) Sort(queues []QueueProperties) {
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

	var firstValue, secondValue = 1, 1

	if first.Error.Has {
		firstValue = 4
	} else if first.Warning.Has {
		firstValue = 2
	}

	if second.Error.Has {
		secondValue = 4
	} else if second.Warning.Has {
		secondValue = 2
	}

	if firstValue > secondValue || first.QueueInfo.MessagesRdy > second.QueueInfo.MessagesRdy {
		return true
	}

	return false
}
