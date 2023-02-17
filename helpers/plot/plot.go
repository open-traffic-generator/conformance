package plot

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/open-traffic-generator/conformance/helpers/table"
)

type DurRange []int64

func (a DurRange) Len() int           { return len(a) }
func (a DurRange) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a DurRange) Less(i, j int) bool { return a[i] < a[j] }

type Distribution struct {
	ApiName  string        `json:"api_name"`
	Duration time.Duration `json:"duration"`
	Type     string
}

type PerApiDistribution struct {
	ApiName         string `json:"api_name"`
	sortedDurations DurRange
	Max             int64 `json:"max"`
	Min             int64 `json:"min"`
	Avg             int64 `json:"avg"`
	P50             int64 `json:"p50"`
	P75             int64 `json:"p75"`
	P90             int64 `json:"p90"`
	P95             int64 `json:"p95"`
	P99             int64 `json:"p99"`
}

type Plot struct {
	Name            string               `json:"name"`
	Iterations      int                  `json:"iterations"`
	Durations       []Duration           `json:"durations"`
	Distributions   []PerApiDistribution `json:"distributions"`
	orderedApiNames []string
}

func NewPlot() *Plot {
	return &Plot{
		Durations: make([]Duration, 0),
	}
}

func (p *Plot) AppendDuration(d Duration) {
	p.Durations = append(p.Durations, d)
}

func (p *Plot) AppendZero() {
	p.Durations = append(p.Durations, Duration{Duration: 0})
}

func (p *Plot) CalculateIterations() {
	p.Iterations = 0

	for i := range p.Durations {
		if p.Durations[i].ApiName == "" {
			p.Iterations += 1
		}
	}
}

func (p *Plot) ApiDurationsMap() map[string]DurRange {
	apiDursMap := map[string]DurRange{}
	p.orderedApiNames = []string{}

	for _, d := range p.Durations {
		if d.ApiName == "" {
			continue
		}

		if _, ok := apiDursMap[d.ApiName]; !ok {
			apiDursMap[d.ApiName] = DurRange{}
			p.orderedApiNames = append(p.orderedApiNames, d.ApiName)
		}

		apiDursMap[d.ApiName] = append(apiDursMap[d.ApiName], int64(d.Duration))
	}

	return apiDursMap
}

func (p *Plot) GetPercentileDuration(durs DurRange, percent int) int64 {
	return durs[(percent*len(durs))/100]
}

func (p *Plot) CalcDistributions() {
	apiDursMap := p.ApiDurationsMap()

	p.Distributions = make([]PerApiDistribution, 0, len(apiDursMap))
	for _, name := range p.orderedApiNames {
		durs := apiDursMap[name]

		sort.Sort(durs)
		d := PerApiDistribution{
			ApiName:         name,
			sortedDurations: durs,
			Min:             durs[0],
			Max:             durs[len(durs)-1],
			Avg: func() int64 {
				sum := int64(0)
				for _, dur := range durs {
					sum += dur
				}

				return sum / int64(len(durs))
			}(),
			P50: p.GetPercentileDuration(durs, 50),
			P75: p.GetPercentileDuration(durs, 75),
			P90: p.GetPercentileDuration(durs, 90),
			P95: p.GetPercentileDuration(durs, 95),
			P99: p.GetPercentileDuration(durs, 99),
		}

		p.Distributions = append(p.Distributions, d)
	}
}

func (p *Plot) Analyze(name string) {
	p.Name = name
	p.CalculateIterations()
	p.CalcDistributions()
}

func (p *Plot) ToJson() (string, error) {
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (p *Plot) ToTable() (string, error) {
	tb := table.NewTable(
		fmt.Sprintf("Distribution: %s (Iterations %d)", p.Name, p.Iterations),
		func() []string {
			cols := make([]string, 0, len(p.Distributions)+1)
			cols = append(cols, "Dist")
			for _, d := range p.Distributions {
				cols = append(cols, d.ApiName)
			}

			return cols
		}(),
		20,
	)

	tb.AppendRow((func() []interface{} {
		row := make([]interface{}, 0, len(p.Distributions)+1)
		row = append(row, "min")
		for _, d := range p.Distributions {
			row = append(row, d.Min)
		}
		return row
	})())

	tb.AppendRow((func() []interface{} {
		row := make([]interface{}, 0, len(p.Distributions)+1)
		row = append(row, "avg")
		for _, d := range p.Distributions {
			row = append(row, d.Avg)
		}
		return row
	})())

	tb.AppendRow((func() []interface{} {
		row := make([]interface{}, 0, len(p.Distributions)+1)
		row = append(row, "max")
		for _, d := range p.Distributions {
			row = append(row, d.Max)
		}
		return row
	})())

	tb.AppendRow((func() []interface{} {
		row := make([]interface{}, 0, len(p.Distributions)+1)
		row = append(row, "p50")
		for _, d := range p.Distributions {
			row = append(row, d.P50)
		}
		return row
	})())

	tb.AppendRow((func() []interface{} {
		row := make([]interface{}, 0, len(p.Distributions)+1)
		row = append(row, "p75")
		for _, d := range p.Distributions {
			row = append(row, d.P75)
		}
		return row
	})())

	tb.AppendRow((func() []interface{} {
		row := make([]interface{}, 0, len(p.Distributions)+1)
		row = append(row, "p90")
		for _, d := range p.Distributions {
			row = append(row, d.P90)
		}
		return row
	})())

	tb.AppendRow((func() []interface{} {
		row := make([]interface{}, 0, len(p.Distributions)+1)
		row = append(row, "p95")
		for _, d := range p.Distributions {
			row = append(row, d.P95)
		}
		return row
	})())

	tb.AppendRow((func() []interface{} {
		row := make([]interface{}, 0, len(p.Distributions)+1)
		row = append(row, "p99")
		for _, d := range p.Distributions {
			row = append(row, d.P99)
		}
		return row
	})())

	return tb.String(), nil
}
