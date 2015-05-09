package rabbitmonit

import (
	"strconv"

	"github.com/c-datculescu/rabbit-hole"
)

/*
QueueProperties offers a broader set of operations than rabbithole.QueueInfo including warnings,
errors and statistics
*/
type QueueProperties struct {
	Stats     QueueStat
	Error     QueueAlert
	Warning   QueueAlert
	QueueInfo rabbithole.QueueInfo
	Client    *rabbithole.Client
}

/*
QueueStat holds a set of queue related statistics
*/
type QueueStat struct {
	NonPersistentMessagesCount int // the number of non-persistent messages in the queue
}

/*
QueueAlert holds various alert flags
*/
type QueueAlert struct {
	State         bool // the status of the queue. will be true if the state is not "running"
	NonDurable    bool // the durability of the queue. will be true if the queue is not durable (will not survive a server restart)
	Rdy           bool // the number of rady messages. 1-100 = warning, >100 = error
	Unack         bool // the number of unacknowledged messages. > ∑ consumer qos = error
	Listener      bool // the number of consumers. 1-3 = warning, 0 = error
	Utilisation   bool // the consumer utilisation. < 70 = warning, < 30 = error
	Intake        bool // the diff between in and out. ∑ in, out > 1 = warning
	NonDurableMsg bool // the number of non-durable messages. will be true if the number of non-durable messages > 1
	Has           bool // identifies whether we have errors/warnings at all
}

/*
Calculate performs various additional calculations based on the details provided by the api
for the queues, initialising and calculating the warnings, alerts and stats
*/
func (qp *QueueProperties) Calculate() {
	qp.Stats = QueueStat{}
	qp.Error = QueueAlert{}
	qp.Warning = QueueAlert{}

	qp.alertState().
		alertDurable().
		alertRdy().
		alertListener().
		alertUtilisation().
		alertIntake().
		alertNonDurableMessages().
		alertUnackMessages()
}

/*
alertRdy calculates whether it should raise an alert or warning when there are messages in the queue that are in
ready status
threshold for warnings is 0
threshold for error is 100
*/
func (qp *QueueProperties) alertRdy() *QueueProperties {
	if qp.QueueInfo.MessagesRdy > 100 {
		qp.Error.Rdy = true
		qp.Error.Has = true
	} else if qp.QueueInfo.MessagesRdy > 0 {
		qp.Warning.Rdy = true
		qp.Warning.Has = true
	}
	return qp
}

/*
alertListener calculates whether there should be an alert/warning raised for the number of listeners on the current
queue
threshold for alert is 0 listeners and more than 0 ready messages
threshold for warning is 3 listeners and more than 0 ready messages
*/
func (qp *QueueProperties) alertListener() *QueueProperties {
	if qp.QueueInfo.Consumers == 0 && qp.QueueInfo.MessagesRdy > 0 {
		qp.Error.Listener = true
		qp.Error.Has = true
	} else if qp.QueueInfo.Consumers <= 3 && qp.QueueInfo.MessagesRdy > 0 {
		qp.Warning.Has = true
		qp.Warning.Listener = true
	}
	return qp
}

/*
alertUtilisation calculates whether there should be an alert/warning raised for the utilisation on the current queue
utilisation is important to give an overview of how many messages end up being dispatched
utlisation is only usable though for high traffic queues
threshold for alert is utilisation less than 30 and more than 0 messages ready in the queue
threshold for warning is utilisation less than 70 and messages ready in the queue
*/
func (qp *QueueProperties) alertUtilisation() *QueueProperties {
	var consumerUtilisation float64
	switch qp.QueueInfo.ConsumerUtilisation.(type) {
	case string:
		if qp.QueueInfo.ConsumerUtilisation == "" {
			qp.QueueInfo.ConsumerUtilisation = "0"
		}
		consumerUtilisation, _ = strconv.ParseFloat(qp.QueueInfo.ConsumerUtilisation.(string), 64)
	case float64:
		consumerUtilisation = qp.QueueInfo.ConsumerUtilisation.(float64)
	}

	if consumerUtilisation < 30 && qp.QueueInfo.MessagesRdy > 0 {
		qp.Error.Has = true
		qp.Error.Utilisation = true
	} else if consumerUtilisation < 70 && qp.QueueInfo.MessagesRdy > 0 {
		qp.Warning.Has = true
		qp.Warning.Utilisation = true
	}
	return qp
}

/*
alertIntake should be deprecated and replaced by a diff between enqueue rate and dequeue rate
@todo replace intake alert with enqueue/dequeue rate difference
*/
func (qp *QueueProperties) alertIntake() *QueueProperties {
	if qp.QueueInfo.MessagesRdyDetails.Rate > 1 {
		qp.Error.Has = true
		qp.Warning.Intake = true
	}
	return qp
}

/*
alertNonDurableMessages raises an alert if the queue contains non-durable messages.
non-durable messages can be lost on a server restart
*/
func (qp *QueueProperties) alertNonDurableMessages() *QueueProperties {
	qp.Stats.NonPersistentMessagesCount = qp.QueueInfo.MessagesRam - qp.QueueInfo.MessagesPersistent
	if qp.Stats.NonPersistentMessagesCount > 0 {
		qp.Error.Has = true
		qp.Error.NonDurableMsg = true
	}
	return qp
}

/*
alertUnackMessages raises an alert by analysing individual properties of consumers on queue as well as unack messages
the threshold for alert is sum of consumer prefetch count is lower than the number of unack messages in the queue
*/
func (qp *QueueProperties) alertUnackMessages() *QueueProperties {
	consumers, err := qp.Client.ConsumersIn(qp.QueueInfo.Vhost)
	if err != nil {
		return qp
	}
	var total int
	for _, consumer := range consumers {
		if qp.QueueInfo.Name == consumer.Queue.Name && qp.QueueInfo.Vhost == consumer.Queue.Vhost {
			total += consumer.PrefetchCount
		}
	}

	if qp.QueueInfo.MessagesUnack > total {
		qp.Error.Has = true
		qp.Error.Unack = true
	}
	return qp
}

/*
alertState raises an alert if the state of the queue is not "running"
*/
func (qp *QueueProperties) alertState() *QueueProperties {
	if qp.QueueInfo.State != "running" {
		qp.Error.Has = true
		qp.Error.State = true
	}
	return qp
}

/*
alertDurable raises an alert if the queue is not durable
*/
func (qp *QueueProperties) alertDurable() *QueueProperties {
	if !qp.QueueInfo.Durable {
		qp.Error.Has = true
		qp.Error.NonDurable = true
	}
	return qp
}
