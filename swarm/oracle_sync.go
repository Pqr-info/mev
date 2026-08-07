package main

type OracleSyncEngine struct {
	firehose *AlpacaFirehose
	tso      *TemporalStabilityOracle
}

func NewOracleSyncEngine(f *AlpacaFirehose, tso *TemporalStabilityOracle) *OracleSyncEngine {
	return &OracleSyncEngine{
		firehose: f,
		tso:      tso,
	}
}

func (o *OracleSyncEngine) SyncEvent(ev TemporalEvent) error {
	report := o.tso.Evaluate(ev.Epoch)

	rec := AlpacaRecord{
		EventID:     ev.EventID,
		PayloadType: ev.PayloadType,
		Risk:        ev.Drift,
		Stability:   report.StabilityClass,
		Epoch:       ev.Epoch,
		Timestamp:   ev.Timestamp,
	}

	return o.firehose.Emit(rec)
}
