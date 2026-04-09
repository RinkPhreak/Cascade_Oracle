package observability

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"cascade/internal/application/port"
)

type MemoryMonitor struct {
	cache port.Cache
}

func NewMemoryMonitor(cache port.Cache) *MemoryMonitor {
	return &MemoryMonitor{cache: cache}
}

// Start initiates the background goroutine reading cgroup stats
func (m *MemoryMonitor) Start(ctx context.Context, thresholdPercent float64) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.evaluateMemory(ctx, thresholdPercent)
		}
	}
}

func (m *MemoryMonitor) evaluateMemory(ctx context.Context, thresholdPercent float64) {
	currentBytes, maxBytes, err := m.getCgroupMemory()
	if err != nil {
		currentBytes, maxBytes = m.getFallbackMemory()
	}

	if maxBytes == 0 {
		return // Cannot evaluate without a limit bound properly
	}

	usage := float64(currentBytes) / float64(maxBytes) * 100.0

	if usage >= thresholdPercent {
		slog.Warn("memory critical threshold breached", "usage", usage, "current_bytes", currentBytes, "max_bytes", maxBytes)
		// Set suspension state for 10 seconds (slightly longer than ticker rhythm)
		m.cache.Set(ctx, "cascade:memory:critical", "1", 10*time.Second) 
	}
}

func (m *MemoryMonitor) getCgroupMemory() (uint64, uint64, error) {
	currentRaw, err := os.ReadFile("/sys/fs/cgroup/memory.current")
	if err != nil {
		return 0, 0, err
	}
	
	maxRaw, err := os.ReadFile("/sys/fs/cgroup/memory.max")
	if err != nil {
		return 0, 0, err
	}

	currentStr := strings.TrimSpace(string(currentRaw))
	maxStr := strings.TrimSpace(string(maxRaw))

	if maxStr == "max" {
		return 0, 0, fmt.Errorf("cgroup limit is set to max")
	}

	currentBytes, err := strconv.ParseUint(currentStr, 10, 64)
	if err != nil {
		return 0, 0, err
	}

	maxBytes, err := strconv.ParseUint(maxStr, 10, 64)
	if err != nil {
		return 0, 0, err
	}

	return currentBytes, maxBytes, nil
}

func (m *MemoryMonitor) getFallbackMemory() (uint64, uint64) {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	used := ms.Sys

	maxStr := os.Getenv("MEMORY_LIMIT_BYTES_FALLBACK")
	maxBytes, _ := strconv.ParseUint(maxStr, 10, 64)

	return used, maxBytes
}
