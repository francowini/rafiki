package objectivebus

// CalculateProgressPercentage calculates the progress percentage for a result-type objective.
// It uses BeginMetric as the starting point (defaults to 0 if not set).
// The direction is auto-detected:
// - If begin < target: increasing goal (e.g., read 35 books)
// - If begin > target: decreasing goal (e.g., lose weight from 80kg to 70kg)
func CalculateProgressPercentage(obj Objective) int {
	if !obj.TrackingType.IsResult() || obj.TargetMetric == nil || obj.CurrentMetric == nil {
		return 0
	}

	target := obj.TargetMetric.Value()
	current := obj.CurrentMetric.Value()
	begin := 0
	if obj.BeginMetric != nil {
		begin = obj.BeginMetric.Value()
	}

	// Auto-detect direction based on begin vs target
	if begin < target {
		// Increasing goal (e.g., read 35 books, run 100km)
		if current <= begin {
			return 0
		}
		if current >= target {
			return 100
		}
		return ((current - begin) * 100) / (target - begin)
	} else if begin > target {
		// Decreasing goal (e.g., weight loss from 80kg to 70kg)
		if current >= begin {
			return 0
		}
		if current <= target {
			return 100
		}
		return ((begin - current) * 100) / (begin - target)
	}

	// begin == target (invalid configuration, but handle gracefully)
	return 0
}
