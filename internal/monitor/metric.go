package monitor

import (
	"fmt"
)

type MetricsByTopic map[string]Metric

func (m MetricsByTopic) String() string {
	var s string
	for k, v := range m {
		s = s + fmt.Sprintf("\t%s >\tsuccesses: %d\terrors: %d\ttotal processed msgs: %d\n", k, v.Successes, v.Errs, v.MsgCount)
	}
	return s
}

type Metric struct {
	Successes int
	Errs      int
	MsgCount  int
}

func SumMetrics(metrics ...Metric) Metric {
	var metric Metric
	for _, mtrc := range metrics {
		metric.Successes += mtrc.Successes
		metric.Errs += mtrc.Errs
		metric.MsgCount += mtrc.MsgCount
	}
	return metric
}
