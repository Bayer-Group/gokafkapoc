package stream

import (
	"context"

	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"github.com/MonsantoCo/gokafkapoc/internal/monitor"
)

type Writer interface {
	WriteMessages(context.Context, ...kafka.Message) error
}

func ProcessMessages(ctx context.Context, w Writer, messages <-chan kafka.Message, topic string) monitor.Metric {
	var metric monitor.Metric
	closer := make(chan struct{})
	msgs := make([]kafka.Message, 0)

	go func() {
		defer close(closer)
		for msg := range messages {
			select {
			case <-ctx.Done():
				return
			default:
				msgs = append(msgs, msg)
				metric.MsgCount++
				if writeMetric, submitted := writeMessageBatch(ctx, w, msgs, topic); submitted {
					msgs = make([]kafka.Message, 0)
					metric = monitor.SumMetrics(metric, writeMetric)
				}
			}
		}
	}()
	<-closer
	return metric
}

func writeMessageBatch(ctx context.Context, w Writer, msgs []kafka.Message, topic string) (monitor.Metric, bool) {
	var (
		submitted bool
		metric    monitor.Metric
	)
	if len(msgs) > 50 {
		if err := w.WriteMessages(ctx, msgs...); err != nil {
			metric.Errs++
			log.Printf("%s: failed to send message: %s", topic, err)
		} else {
			submitted = true
			metric.Successes++
			log.Printf("> %s: successful batch", topic)
		}
	}
	return metric, submitted
}
