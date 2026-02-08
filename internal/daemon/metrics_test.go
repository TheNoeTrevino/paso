package daemon

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetrics(t *testing.T) {
	m := NewMetrics()

	require.NotNil(t, m)

	// Verify all counters start at zero
	assert.Equal(t, int64(0), m.GetEventsSent())
	assert.Equal(t, int64(0), m.GetEventsReceived())
	assert.Equal(t, int64(0), m.GetReconnections())
	assert.Equal(t, int64(0), m.GetRefreshesTotal())
	assert.Equal(t, int32(0), m.GetConnectedClients())

	// Verify StartTime is set to a recent time (within last second)
	assert.WithinDuration(t, time.Now(), m.StartTime, time.Second)

	t.Logf("Metrics initialized correctly: %+v", m.GetSnapshot())
}

func TestIncEventsSent(t *testing.T) {
	m := NewMetrics()

	// Increment once
	m.IncEventsSent()
	assert.Equal(t, int64(1), m.GetEventsSent())

	// Increment multiple times
	for i := 0; i < 10; i++ {
		m.IncEventsSent()
	}
	assert.Equal(t, int64(11), m.GetEventsSent())

	t.Logf("EventsSent incremented correctly: %d", m.GetEventsSent())
}

func TestIncEventsReceived(t *testing.T) {
	m := NewMetrics()

	m.IncEventsReceived()
	assert.Equal(t, int64(1), m.GetEventsReceived())

	for i := 0; i < 5; i++ {
		m.IncEventsReceived()
	}
	assert.Equal(t, int64(6), m.GetEventsReceived())

	t.Logf("EventsReceived incremented correctly: %d", m.GetEventsReceived())
}

func TestIncReconnections(t *testing.T) {
	m := NewMetrics()

	m.IncReconnections()
	assert.Equal(t, int64(1), m.GetReconnections())

	for i := 0; i < 3; i++ {
		m.IncReconnections()
	}
	assert.Equal(t, int64(4), m.GetReconnections())

	t.Logf("Reconnections incremented correctly: %d", m.GetReconnections())
}

func TestIncRefreshesTotal(t *testing.T) {
	m := NewMetrics()

	m.IncRefreshesTotal()
	assert.Equal(t, int64(1), m.GetRefreshesTotal())

	for i := 0; i < 20; i++ {
		m.IncRefreshesTotal()
	}
	assert.Equal(t, int64(21), m.GetRefreshesTotal())

	t.Logf("RefreshesTotal incremented correctly: %d", m.GetRefreshesTotal())
}

func TestSetConnectedClients(t *testing.T) {
	m := NewMetrics()

	// Set to various values
	m.SetConnectedClients(5)
	assert.Equal(t, int32(5), m.GetConnectedClients())

	m.SetConnectedClients(0)
	assert.Equal(t, int32(0), m.GetConnectedClients())

	m.SetConnectedClients(100)
	assert.Equal(t, int32(100), m.GetConnectedClients())

	t.Logf("ConnectedClients set correctly: %d", m.GetConnectedClients())
}

func TestGetSnapshot(t *testing.T) {
	m := NewMetrics()

	// Set some values
	m.IncEventsSent()
	m.IncEventsSent()
	m.IncEventsReceived()
	m.IncReconnections()
	m.IncRefreshesTotal()
	m.SetConnectedClients(3)

	// Give it a moment so uptime is measurable
	time.Sleep(10 * time.Millisecond)

	snapshot := m.GetSnapshot()

	// Verify all fields
	assert.Equal(t, int64(2), snapshot.EventsSent)
	assert.Equal(t, int64(1), snapshot.EventsReceived)
	assert.Equal(t, int64(1), snapshot.Reconnections)
	assert.Equal(t, int64(1), snapshot.RefreshesTotal)
	assert.Equal(t, int32(3), snapshot.ConnectedClients)

	// Verify StartTime matches
	assert.True(t, snapshot.StartTime.Equal(m.StartTime))

	// Verify Uptime is populated and reasonable
	assert.NotZero(t, snapshot.Uptime)
	t.Logf("Uptime: %s", snapshot.Uptime)

	// Verify uptime is at least the sleep duration
	expectedUptime := 10 * time.Millisecond
	actualUptime := time.Since(m.StartTime)
	assert.True(t, actualUptime >= expectedUptime)

	t.Logf("Snapshot captured correctly: %+v", snapshot)
}

func TestMetricsConcurrency_AllOperations(t *testing.T) {
	m := NewMetrics()

	// Number of goroutines and operations per goroutine
	numGoroutines := 100
	opsPerGoroutine := 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 5) // 5 different operations

	// Concurrently increment EventsSent
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				m.IncEventsSent()
			}
		}()
	}

	// Concurrently increment EventsReceived
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				m.IncEventsReceived()
			}
		}()
	}

	// Concurrently increment Reconnections
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				m.IncReconnections()
			}
		}()
	}

	// Concurrently increment RefreshesTotal
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				m.IncRefreshesTotal()
			}
		}()
	}

	// Concurrently set ConnectedClients
	for i := 0; i < numGoroutines; i++ {
		go func(val int32) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				m.SetConnectedClients(val)
			}
		}(int32(i))
	}

	wg.Wait()

	// Verify counts are correct
	expectedCount := int64(numGoroutines * opsPerGoroutine)
	assert.Equal(t, expectedCount, m.GetEventsSent())
	assert.Equal(t, expectedCount, m.GetEventsReceived())
	assert.Equal(t, expectedCount, m.GetReconnections())
	assert.Equal(t, expectedCount, m.GetRefreshesTotal())

	// ConnectedClients is set (not incremented), so it should be one of the values
	clientCount := m.GetConnectedClients()
	assert.True(t, clientCount >= 0 && clientCount < int32(numGoroutines))

	t.Logf("Concurrent operations completed successfully")
	t.Logf("  Final counts: EventsSent=%d, EventsReceived=%d, Reconnections=%d, RefreshesTotal=%d, ConnectedClients=%d",
		m.GetEventsSent(), m.GetEventsReceived(), m.GetReconnections(), m.GetRefreshesTotal(), m.GetConnectedClients())
}

func TestMetricsConcurrency_ReadWhileWrite(t *testing.T) {
	m := NewMetrics()

	stopChan := make(chan struct{})
	var wg sync.WaitGroup

	// Start writers
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stopChan:
					return
				default:
					m.IncEventsSent()
					m.IncEventsReceived()
					m.IncRefreshesTotal()
					m.SetConnectedClients(5)
				}
			}
		}()
	}

	// Start readers
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stopChan:
					return
				default:
					_ = m.GetEventsSent()
					_ = m.GetEventsReceived()
					_ = m.GetRefreshesTotal()
					_ = m.GetConnectedClients()
					_ = m.GetSnapshot()
				}
			}
		}()
	}

	// Run for 100ms
	time.Sleep(100 * time.Millisecond)
	close(stopChan)
	wg.Wait()

	snapshot := m.GetSnapshot()
	t.Logf("Concurrent read/write operations completed successfully")
	t.Logf("  Final snapshot: %+v", snapshot)

	// Verify metrics are reasonable (non-negative, etc.)
	assert.False(t, snapshot.EventsSent < 0)
	assert.False(t, snapshot.EventsReceived < 0)
	assert.False(t, snapshot.RefreshesTotal < 0)
}

func TestMetricsSnapshot_IsImmutable(t *testing.T) {
	m := NewMetrics()

	m.IncEventsSent()
	snapshot1 := m.GetSnapshot()

	// Change metrics after taking snapshot
	m.IncEventsSent()
	m.IncEventsSent()

	// Verify snapshot hasn't changed
	assert.Equal(t, int64(1), snapshot1.EventsSent)

	// Take another snapshot
	snapshot2 := m.GetSnapshot()
	assert.Equal(t, int64(3), snapshot2.EventsSent)

	t.Logf("Snapshots are immutable and independent")
}
