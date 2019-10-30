package utilities

import (
	"errors"
	"fmt"

	dto "github.com/prometheus/client_model/go"
)

// GetCounter Get a counter metric value from the provided metric families
func GetCounter(name string, families []*dto.MetricFamily) (value float64, err error) {

	hasMetric := false
	for _, family := range families {
		metrics := family.GetMetric()

		if len(metrics) == 0 {
			return 0, errors.New("No metrics found in Family")
		}

		for _, m := range metrics {
			c := m.GetCounter()
			if family.GetName() == name {
				value += c.GetValue()
				hasMetric = true
			}
		}
	}
	if hasMetric {
		return value, nil
	}

	return 0, fmt.Errorf("Unable to find counter named %s", name)
}

// GetHistogramSum Get a histogram metric sum from the provided metric families
func GetHistogramSum(name string, families []*dto.MetricFamily) (value float64, err error) {

	hasMetric := false
	for _, family := range families {
		metrics := family.GetMetric()

		if len(metrics) == 0 {
			return 0, errors.New("No metrics found in Family")
		}

		for _, m := range metrics {
			h := m.GetHistogram()
			if family.GetName() == name {
				value += h.GetSampleSum()
				hasMetric = true
			}
		}
	}
	if hasMetric {
		return value, nil
	}

	return 0, fmt.Errorf("Unable to find counter named %s", name)
}
