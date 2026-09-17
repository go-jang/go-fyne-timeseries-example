package internal

import (
	"context"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/go-errr/go/err"
	"github.com/go-fyne-timeseries/go/chart"
)

type TimeSeriesProvider struct {
	mu         sync.RWMutex
	timeSeries map[string]*chart.TimeSeries
}

func NewTimeSeriesProvider() *TimeSeriesProvider {
	return &TimeSeriesProvider{
		timeSeries: make(map[string]*chart.TimeSeries),
	}
}

func (this *TimeSeriesProvider) register(dataset *chart.TimeSeriesCollection) {
	this.mu.Lock()
	defer this.mu.Unlock()
	dataset.Data().ForEach(func(ts *chart.TimeSeries) {
		if _, exists := this.timeSeries[ts.Name()]; !exists {
			this.timeSeries[ts.Name()] = ts
		}
	})
}

func (this *TimeSeriesProvider) get(id int64) *chart.TimeSeries {
	this.mu.RLock()
	defer this.mu.RUnlock()
	return this.timeSeries[strconv.FormatInt(id, 10)]
}

func (this *TimeSeriesProvider) Subscribe(ctx context.Context, paramIds []int64, dataset *chart.TimeSeriesCollection) {
	this.register(dataset)
	go func() {
		defer err.Recover()
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				for _, id := range paramIds {
					if !this.mockPresent(id, now) {
						continue
					}
					if ts := this.get(id); ts != nil {
						ts.Add(chart.TimeSeriesDataItemOf(float64(now.UnixMilli()), this.mockValue(id, now)), true)
					}
				}
			}
		}
	}()
}

func (this *TimeSeriesProvider) GetHistoricalDataAsync(ctx context.Context, paramIds []int64, fromInclusive, toExclusive time.Time, dataset *chart.TimeSeriesCollection) {
	this.register(dataset)
	go func() {
		defer err.Recover()
		for t := fromInclusive; t.Before(toExclusive); t = t.Add(time.Second) {
			if ctx.Err() != nil {
				return
			}
			for _, id := range paramIds {
				if !this.mockPresent(id, t) {
					continue
				}
				if ts := this.get(id); ts != nil {
					ts.Add(chart.TimeSeriesDataItemOf(float64(t.UnixMilli()), this.mockValue(id, t)), true)
				}
			}
		}
	}()
}

func (this *TimeSeriesProvider) mockValue(id int64, t time.Time) float64 {
	seconds := float64(t.UnixMilli()) / 1000
	phase := float64(id) * 0.7
	value := 50 +
		float64(id%5)*15 +
		20*math.Sin(seconds/30+phase) +
		5*math.Sin(seconds/3+phase)
	// slow trend
	value += 0.002 * math.Mod(seconds, 3600)
	// periodic step change
	if int64(seconds/120)%2 == 1 {
		value += 15
	}
	// occasional single-sample spike
	if int64(seconds+float64(id*7))%137 == 0 {
		value += 100
	}
	return value
}

func (this *TimeSeriesProvider) mockPresent(id int64, t time.Time) bool {
	seconds := t.Unix()

	// 10-second gap approximately every 3 minutes,
	// shifted per series so gaps don't coincide.
	position := (seconds + id*17) % 180
	return position >= 10
}
