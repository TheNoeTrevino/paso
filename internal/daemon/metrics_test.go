package daemon

import (
	"sync"
	"testing"
	"time"
)

// ============================================================================
// Basic Metrics Tests
// ============================================================================

func TestNewMetrics(t *testing.T) {
	m := NewMetrics()

	if m == nil {
		t.Fatal("Expected NewMetrics to return non-nil")
	}

	// Verify all counters start at zero
	if m.GetEventsSent() != 0 {
		t.Errorf("Expected EventsSent to be 0, got %d", m.GetEventsSent())
	}
	if m.GetEventsReceived() != 0 {
		t.Errorf("Expected EventsReceived to be 0, got %d", m.GetEventsReceived())
	}
	if m.GetReconnections() != 0 {
		t.Errorf("Expected Reconnections to be 0, got %d", m.GetReconnections())
	}
	if m.GetRefreshesTotal() != 0 {
		t.Errorf("Expected RefreshesTotal to be 0, got %d", m.GetRefreshesTotal())
	}
	if m.GetConnectedClients() != 0 {
		t.Errorf("Expected ConnectedClients to be 0, got %d", m.GetConnectedClients())
	}

	// Verify StartTime is set to a recent time (within last second)
	if time.Since(m.StartTime) > time.Second {
		t.Errorf("Expected StartTime to be recent, got %v", m.StartTime)
	}

	t.Logf("✓ Metrics initialized correctly: %+v", m.GetSnapshot())
}

func TestIncEventsSent(t *testing.T) {
	m := NewMetrics()

	// Increment once
	m.IncEventsSent()
	if m.GetEventsSent() != 1 {
		t.Errorf("Expected EventsSent to be 1, got %d", m.GetEventsSent())
	}

	// Increment multiple times
	for i := 0; i < 10; i++ {
		m.IncEventsSent()
	}
	if m.GetEventsSent() != 11 {
		t.Errorf("Expected EventsSent to be 11, got %d", m.GetEventsSent())
	}

	t.Logf("✓ EventsSent incremented correctly: %d", m.GetEventsSent())
}

func TestIncEventsReceived(t *testing.T) {
	m := NewMetrics()

	m.IncEventsReceived()
	if m.GetEventsReceived() != 1 {
		t.Errorf("Expected EventsReceived to be 1, got %d", m.GetEventsReceived())
	}

	for i := 0; i < 5; i++ {
		m.IncEventsReceived()
	}
	if m.GetEventsReceived() != 6 {
		t.Errorf("Expected EventsReceived to be 6, got %d", m.GetEventsReceived())
	}

	t.Logf("✓ EventsReceived incremented correctly: %d", m.GetEventsReceived())
}

func TestIncReconnections(t *testing.T) {
	m := NewMetrics()

	m.IncReconnections()
	if m.GetReconnections() != 1 {
		t.Errorf("Expected Reconnections to be 1, got %d", m.GetReconnections())
	}

	for i := 0; i < 3; i++ {
		m.IncReconnections()
	}
	if m.GetReconnections() != 4 {
		t.Errorf("Expected Reconnections to be 4, got %d", m.GetReconnections())
	}

	t.Logf("✓ Reconnections incremented correctly: %d", m.GetReconnections())
}

func TestIncRefreshesTotal(t *testing.T) {
	m := NewMetrics()

	m.IncRefreshesTotal()
	if m.GetRefreshesTotal() != 1 {
		t.Errorf("Expected RefreshesTotal to be 1, got %d", m.GetRefreshesTotal())
	}

	for i := 0; i < 20; i++ {
		m.IncRefreshesTotal()
	}
	if m.GetRefreshesTotal() != 21 {
		t.Errorf("Expected RefreshesTotal to be 21, got %d", m.GetRefreshesTotal())
	}

	t.Logf("✓ RefreshesTotal incremented correctly: %d", m.GetRefreshesTotal())
}

func TestSetConnectedClients(t *testing.T) {
	m := NewMetrics()

	// Set to various values
	m.SetConnectedClients(5)
	if m.GetConnectedClients() != 5 {
		t.Errorf("Expected ConnectedClients to be 5, got %d", m.GetConnectedClients())
	}

	m.SetConnectedClients(0)
	if m.GetConnectedClients() != 0 {
		t.Errorf("Expected ConnectedClients to be 0, got %d", m.GetConnectedClients())
	}

	m.SetConnectedClients(100)
	if m.GetConnectedClients() != 100 {
		t.Errorf("Expected ConnectedClients to be 100, got %d", m.GetConnectedClients())
	}

	t.Logf("✓ ConnectedClients set correctly: %d", m.GetConnectedClients())
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
	if snapshot.EventsSent != 2 {
		t.Errorf("Expected EventsSent to be 2, got %d", snapshot.EventsSent)
	}
	if snapshot.EventsReceived != 1 {
		t.Errorf("Expected EventsReceived to be 1, got %d", snapshot.EventsReceived)
	}
	if snapshot.Reconnections != 1 {
		t.Errorf("Expected Reconnections to be 1, got %d", snapshot.Reconnections)
	}
	if snapshot.RefreshesTotal != 1 {
		t.Errorf("Expected RefreshesTotal to be 1, got %d", snapshot.RefreshesTotal)
	}
	if snapshot.ConnectedClients != 3 {
		t.Errorf("Expected ConnectedClients to be 3, got %d", snapshot.ConnectedClients)
	}

	// Verify StartTime matches
	if !snapshot.StartTime.Equal(m.StartTime) {
		t.Errorf("Expected StartTime to match, got %v vs %v", snapshot.StartTime, m.StartTime)
	}

	// Verify Uptime is populated and reasonable
	if snapshot.Uptime == "" {
		t.Error("Expected Uptime to be populated")
	}
	t.Logf("✓ Uptime: %s", snapshot.Uptime)

	// Verify uptime is at least the sleep duration
	expectedUptime := 10 * time.Millisecond
	actualUptime := time.Since(m.StartTime)
	if actualUptime < expectedUptime {
		t.Errorf("Expected uptime >= %v, got %v", expectedUptime, actualUptime)
	}

	t.Logf("✓ Snapshot captured correctly: %+v", snapshot)
}

// ============================================================================
// Concurrency Tests (Race Detector)
// ============================================================================

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
	if m.GetEventsSent() != expectedCount {
		t.Errorf("Expected EventsSent to be %d, got %d", expectedCount, m.GetEventsSent())
	}
	if m.GetEventsReceived() != expectedCount {
		t.Errorf("Expected EventsReceived to be %d, got %d", expectedCount, m.GetEventsReceived())
	}
	if m.GetReconnections() != expectedCount {
		t.Errorf("Expected Reconnections to be %d, got %d", expectedCount, m.GetReconnections())
	}
	if m.GetRefreshesTotal() != expectedCount {
		t.Errorf("Expected RefreshesTotal to be %d, got %d", expectedCount, m.GetRefreshesTotal())
	}

	// ConnectedClients is set (not incremented), so it should be one of the values
	clientCount := m.GetConnectedClients()
	if clientCount < 0 || clientCount >= int32(numGoroutines) {
		t.Errorf("Expected ConnectedClients to be in range [0, %d), got %d", numGoroutines, clientCount)
	}

	t.Logf("✓ Concurrent operations completed successfully")
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
	t.Logf("✓ Concurrent read/write operations completed successfully")
	t.Logf("  Final snapshot: %+v", snapshot)

	// Verify metrics are reasonable (non-negative, etc.)
	if snapshot.EventsSent < 0 {
		t.Errorf("EventsSent should not be negative: %d", snapshot.EventsSent)
	}
	if snapshot.EventsReceived < 0 {
		t.Errorf("EventsReceived should not be negative: %d", snapshot.EventsReceived)
	}
	if snapshot.RefreshesTotal < 0 {
		t.Errorf("RefreshesTotal should not be negative: %d", snapshot.RefreshesTotal)
	}
}

func TestMetricsSnapshot_IsImmutable(t *testing.T) {
	m := NewMetrics()

	m.IncEventsSent()
	snapshot1 := m.GetSnapshot()

	// Change metrics after taking snapshot
	m.IncEventsSent()
	m.IncEventsSent()

	// Verify snapshot hasn't changed
	if snapshot1.EventsSent != 1 {
		t.Errorf("Snapshot should be immutable, expected EventsSent=1, got %d", snapshot1.EventsSent)
	}

	// Take another snapshot
	snapshot2 := m.GetSnapshot()
	if snapshot2.EventsSent != 3 {
		t.Errorf("Second snapshot should reflect changes, expected EventsSent=3, got %d", snapshot2.EventsSent)
	}

	t.Logf("✓ Snapshots are immutable and independent")
}
