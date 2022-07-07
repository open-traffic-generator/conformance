package plot

type Plot struct {
	Durations []Duration
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
