package main

type OracleReplayValidator struct {
	tme *TemporalMemoryEngine
}

func NewOracleReplayValidator(t *TemporalMemoryEngine) *OracleReplayValidator {
	return &OracleReplayValidator{
		tme: t,
	}
}

func (v *OracleReplayValidator) Validate(epoch string, oracleEvents []TemporalEvent) float64 {
	local := v.tme.GetEventsByEpoch(epoch)

	if len(local) == 0 || len(oracleEvents) == 0 {
		return 0.0
	}

	var sum float64
	n := len(local)
	if len(oracleEvents) < n {
		n = len(oracleEvents)
	}

	for i := 0; i < n; i++ {
		d := local[i].Drift - oracleEvents[i].Drift
		if d < 0 {
			d = -d
		}
		sum += d
	}

	return sum / float64(n)
}
