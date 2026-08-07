package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type FirehoseEnvelope struct {
	Source      string            `json:"source"`
	StreamID    string            `json:"stream_id"`
	Timestamp   time.Time         `json:"timestamp"`
	PayloadType string            `json:"payload_type"`
	PayloadRaw  json.RawMessage   `json:"payload"`
	Meta        map[string]string `json:"meta,omitempty"`
}

type MeshEventPayload struct {
	SigmaID    string   `json:"sigma_id"`
	Agent      string   `json:"agent"`
	Files      []string `json:"files"`
	RiskScore  float64  `json:"risk_score"`
	Confidence float64  `json:"confidence"`
}

type MetricPayload struct {
	Name   string            `json:"name"`
	Value  float64           `json:"value"`
	Labels map[string]string `json:"labels,omitempty"`
}

type TeleporterDeserializer interface {
	DeserializeLine(line []byte) (*FirehoseEnvelope, interface{}, error)
	DeserializeStream(r io.Reader, fn func(*FirehoseEnvelope, interface{}) error) error
}

type DefaultTeleporterDeserializer struct {
	registry map[string]func(json.RawMessage) (interface{}, error)
}

func NewDefaultTeleporterDeserializer() *DefaultTeleporterDeserializer {
	return &DefaultTeleporterDeserializer{
		registry: map[string]func(json.RawMessage) (interface{}, error){
			"mesh_event": func(raw json.RawMessage) (interface{}, error) {
				var p MeshEventPayload
				if err := json.Unmarshal(raw, &p); err != nil {
					return nil, err
				}
				return p, nil
			},
			"metric": func(raw json.RawMessage) (interface{}, error) {
				var p MetricPayload
				if err := json.Unmarshal(raw, &p); err != nil {
					return nil, err
				}
				return p, nil
			},
		},
	}
}

func (d *DefaultTeleporterDeserializer) DeserializeLine(line []byte) (*FirehoseEnvelope, interface{}, error) {
	var env FirehoseEnvelope
	if err := json.Unmarshal(line, &env); err != nil {
		return nil, nil, err
	}

	decodeFn, ok := d.registry[env.PayloadType]
	if !ok {
		return &env, nil, fmt.Errorf("unknown payload_type: %s", env.PayloadType)
	}

	payload, err := decodeFn(env.PayloadRaw)
	if err != nil {
		return &env, nil, err
	}

	return &env, payload, nil
}

func (d *DefaultTeleporterDeserializer) DeserializeStream(r io.Reader, fn func(*FirehoseEnvelope, interface{}) error) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		env, payload, err := d.DeserializeLine(line)
		if err != nil {
			return err
		}
		if err := fn(env, payload); err != nil {
			return err
		}
	}
	return scanner.Err()
}
