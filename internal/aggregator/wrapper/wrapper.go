package wrapper

import (
	"encoding/json"
	"fmt"

	"github.com/vahidmostofi/wise-auto-scaler/internal/identifier"

	"github.com/streadway/amqp"
)

type message struct {
	// ID           string  `json:"id"`
	Timestamp    int64   `json:"timestamp"`
	HTTPMethod   string  `json:"httpMethod"`
	Path         string  `json:"path"`
	HTTPCode     int     `json:"httpCode"`
	ResponseTime float64 `json:"responseTime"`
}

// WrapperAggregator ...
type WrapperAggregator struct {
	conn              *amqp.Connection
	ch                *amqp.Channel
	q                 amqp.Queue
	queueName         string
	msgs              <-chan amqp.Delivery
	requestIdentifier identifier.Identifier
	DB                *Database
}

// GetNewWrapperAggregator ..
func GetNewWrapperAggregator() (*WrapperAggregator, error) {
	a := &WrapperAggregator{}
	err := a.init()
	if err != nil {
		return nil, fmt.Errorf("error while initializing WrapperAggregator: %w", err)
	}
	go a.listen()
	return a, nil
}

func (w *WrapperAggregator) init() error {
	const url = "amqp://136.159.209.214:5672"
	const queueName = "monitoring.requests"
	const requestInterval int64 = 1000

	ri, err := identifier.GetNewIdentifier()
	if err != nil {
		return fmt.Errorf("error while creating request identifer: %w", err)
	}
	w.requestIdentifier = ri

	w.DB = GetNewDatabase(w.requestIdentifier.GetTypes(), requestInterval)

	conn, err := amqp.Dial(url)
	w.queueName = queueName
	if err != nil {
		return fmt.Errorf("error while connecting to Rabbitmq at %s: %w", url, err)
	}
	w.conn = conn
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("error while creating channel to Rabbitmq at %s: %w", url, err)
	}
	w.ch = ch
	q, err := ch.QueueDeclare(
		w.queueName, // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		return fmt.Errorf("error while creating queue (%s) for Rabbitmq: %w", w.queueName, err)
	}
	w.q = q
	msgs, err := ch.Consume(
		w.q.Name, // queue
		"",       // consumer
		true,     // auto-ack
		false,    // exclusive
		false,    // no-local
		false,    // no-wait
		nil,      // args
	)
	if err != nil {
		return fmt.Errorf("error while creating consumer from queue (%s) for Rabbitmq: %w", w.queueName, err)
	}
	w.msgs = msgs
	return nil
}

func (w *WrapperAggregator) listen() {
	forever := make(chan bool)

	go func() {
		for d := range w.msgs {
			m := &message{}
			json.Unmarshal(d.Body, m)
			info := make(map[string]string)
			info[identifier.INFOHTTPMethod] = m.HTTPMethod
			info[identifier.INFOHTTPPath] = m.Path
			requestType, err := w.requestIdentifier.GetType(info)
			if err != nil {
				panic(err)
			}
			ri := &RequestInfo{
				ResponseTime: m.ResponseTime,
				Success:      m.HTTPCode < 250,
				Timestamp:    m.Timestamp - int64(m.ResponseTime), //TODO WTF!!
			}
			w.DB.Add(requestType, ri)
		}
	}()

	fmt.Printf(" [*] Waiting for messages. To exit press CTRL+C\n")
	<-forever
	fmt.Println("wrapper exited")
}

// GetRequestsNames ....
func (w *WrapperAggregator) GetRequestsNames(start, end int64) ([]string, error) {
	return w.requestIdentifier.GetTypes(), nil
}

// GetRequestResponseTimes ....
func (w *WrapperAggregator) GetRequestResponseTimes(rt string, start, end int64) ([]float64, error) {
	ris := w.DB.Get(rt, start, end)
	responesTimes := make([]float64, len(ris))
	for i, v := range ris {
		responesTimes[i] = v.ResponseTime
	}
	return responesTimes, nil
}

// GetRequestsResponseTimes ....
func (w *WrapperAggregator) GetRequestsResponseTimes(start, end int64) (map[string][]float64, error) {
	responesTimes := make(map[string][]float64)
	for _, rt := range w.requestIdentifier.GetTypes() {
		r, err := w.GetRequestResponseTimes(rt, start, end)
		if err != nil {
			return nil, err
		}
		responesTimes[rt] = r
	}
	return responesTimes, nil
}

// Done ....
func (w *WrapperAggregator) Done() error { return nil }

// AllDone ....
func (w *WrapperAggregator) AllDone() error {
	err := w.conn.Close()
	if err != nil {
		return err
	}
	err = w.ch.Close()
	if err != nil {
		return err
	}
	return nil
}

// GetRequestCount ....
func (w *WrapperAggregator) GetRequestCount(rt string, start, end int64) (int, error) {
	ris := w.DB.Get(rt, start, end)
	return len(ris), nil
}

// GetRequestsCounts ....
func (w *WrapperAggregator) GetRequestsCounts(start, end int64) (map[string]int, error) {
	counts := make(map[string]int)
	for _, rt := range w.requestIdentifier.GetTypes() {
		c, err := w.GetRequestCount(rt, start, end)
		if err != nil {
			return nil, err
		}
		counts[rt] = c
	}
	return counts, nil
}

// GetFailCount ....
func (w *WrapperAggregator) GetFailCount(rt string, start, end int64) (int, error) {
	ris := w.DB.Get(rt, start, end)
	c := 0
	for _, r := range ris {
		if !r.Success {
			c++
		}
	}
	return c, nil
}

// GetFailCounts ....
func (w *WrapperAggregator) GetFailCounts(start, end int64) (map[string]int, error) {
	failCounts := make(map[string]int)
	for _, rt := range w.requestIdentifier.GetTypes() {
		fc, err := w.GetRequestCount(rt, start, end)
		if err != nil {
			return nil, err
		}
		failCounts[rt] = fc
	}
	return failCounts, nil
}

/*
HOW TO USE?
	fmt.Printf("Received a message: %s, %d, %f\n", requestType, ri.Timestamp, m.ResponseTime)
	fmt.Println("in last 3 seconds: ")
	start := time.Now().UnixNano() / 1e6

	for _, r := range w.requestIdentifier.GetTypes() {
		c, err := w.GetRequestCount(r, start-3*1e3, start)
		if err != nil {
			panic(err)
		}
		rts, err := w.GetRequestResponseTimes(r, start-3*1e3, start)
		if err != nil {
			panic(err)
		}
		fmt.Println(c, rts)
	}

	rts, _ := w.GetRequestsCounts(start-3*1e3, start)
	c, _ := w.GetRequestsResponseTimes(start-3*1e3, start)
	fmt.Println(rts, c)
*/
