package main

import (
	"auth-haven/internal/config"
	"testing"
	"time"
)

type poolRecorder struct {
	maxOpen, maxIdle int
	lifetime         time.Duration
}

func (p *poolRecorder) SetMaxOpenConns(value int)              { p.maxOpen = value }
func (p *poolRecorder) SetMaxIdleConns(value int)              { p.maxIdle = value }
func (p *poolRecorder) SetConnMaxLifetime(value time.Duration) { p.lifetime = value }

func TestConfigureDatabasePoolAppliesLimits(t *testing.T) {
	recorder := &poolRecorder{}
	configureDatabasePool(recorder, config.DatabaseConfig{MaxConnections: 19, MaxIdleConns: 7, ConnMaxLifetime: 3 * time.Minute})
	if recorder.maxOpen != 19 || recorder.maxIdle != 7 || recorder.lifetime != 3*time.Minute {
		t.Fatalf("pool settings = %+v", recorder)
	}
}
