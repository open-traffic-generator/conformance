package otg

import (
	"sort"
	"time"

	"github.com/open-traffic-generator/conformance/helpers/table"
)

type ThroughputMetric struct {
	startTime      time.Time
	txPpsSnapshots []int

	ConfiguredLineSpeedMbps int
	ConfiguredLineRate      float32
	ConfiguredPps           int
	ConfiguredFrames        uint64
	ConfiguredSize          int
	ConfiguredDuration      time.Duration
	TxFrames                uint64
	MinTxPps                int
	MaxTxPps                int
	AvgTxPps                int
	TxThroughput            float32
	TransmitDuration        time.Duration
	Ok                      bool
}

type ThroughputMetrics struct {
	Metrics []ThroughputMetric
}

func NewThroughputMetric(lineSpeedMbps int, lineRate float32, frames uint64, size int) *ThroughputMetric {
	m := &ThroughputMetric{
		ConfiguredLineSpeedMbps: lineSpeedMbps,
		ConfiguredLineRate:      lineRate,
		ConfiguredFrames:        frames,
		ConfiguredSize:          size,
	}

	return m
}

func (m *ThroughputMetric) AddPpsSnapshot(txFrames uint64, txPps int) {
	m.TxFrames = txFrames

	m.txPpsSnapshots = append(m.txPpsSnapshots, txPps)
}

func (m *ThroughputMetric) StartCollecting() {
	bits := (m.ConfiguredSize + 12 + 8) * 8

	m.ConfiguredPps = (m.ConfiguredLineSpeedMbps * int(m.ConfiguredLineRate) * 10000) / bits
	m.ConfiguredDuration = time.Duration((m.ConfiguredFrames * 1000 / uint64(m.ConfiguredPps))) * time.Millisecond
	m.startTime = time.Now()
}

func (m *ThroughputMetric) StopCollecting() {
	m.TransmitDuration = time.Since(m.startTime)

	// remove rates with 0 as PPS
	txPpsSnapshots := []int{}
	for _, v := range m.txPpsSnapshots {
		if v != 0 {
			txPpsSnapshots = append(txPpsSnapshots, v)
		}
	}
	sort.Ints(txPpsSnapshots)

	m.txPpsSnapshots = txPpsSnapshots

	m.MinTxPps = m.txPpsSnapshots[0]
	m.MaxTxPps = m.txPpsSnapshots[len(m.txPpsSnapshots)-1]
	m.AvgTxPps = func() int {
		val := 0
		for _, v := range m.txPpsSnapshots {
			val += v
		}
		return val / len(m.txPpsSnapshots)
	}()

	// keep only two decimal places
	m.TxThroughput = float32(int(float32(m.MaxTxPps)/float32(m.ConfiguredPps)*100)) / 100

	if m.TxFrames == m.ConfiguredFrames {
		m.Ok = true
	}
}

func (m *ThroughputMetrics) ToTable() (string, error) {
	tb := table.NewTable(
		"Throughput Metrics",
		[]string{
			"InSpeedMbps",
			"InLineRate",
			"InSize",
			"TxThroughput",
			"InFrames",
			"TxFrames",
			"InPps",
			"MinTxPps",
			"MaxTxPps",
			"AvgTxPps",
			"InTransmitMs",
			"OutTransmitMs",
			"Ok",
		},
		15,
	)
	for _, v := range m.Metrics {
		tb.AppendRow([]interface{}{
			v.ConfiguredLineSpeedMbps,
			v.ConfiguredLineRate,
			v.ConfiguredSize,
			v.TxThroughput,
			v.ConfiguredFrames,
			v.TxFrames,
			v.ConfiguredPps,
			v.MinTxPps,
			v.MaxTxPps,
			v.AvgTxPps,
			v.ConfiguredDuration.Milliseconds(),
			v.TransmitDuration.Milliseconds(),
			v.Ok,
		})
	}

	return tb.String(), nil
}
