package core

// HistorySize is the limit to how big the history of the relay entries should be
const HistorySize = 6

// HistoryMax returns the max value in the history array
func HistoryMax(history []float32) float32 {
	var max float32
	for i := 0; i < len(history); i++ {
		if history[i] > max {
			max = history[i]
		}
	}
	return max
}

// HistoryNotSet returns a history array initialized with invalid history values
func HistoryNotSet() [HistorySize]float32 {
	var res [HistorySize]float32
	for i := 0; i < HistorySize; i++ {
		res[i] = InvalidHistoryValue
	}
	return res
}

// HistoryMean returns the average value of all the history entries
func HistoryMean(history []float32) float32 {
	var sum float32
	var size int
	for i := 0; i < len(history); i++ {
		if history[i] != InvalidHistoryValue {
			sum += history[i]
			size++
		}
	}
	if size == 0 {
		return 0
	}
	return sum / float32(size)
}
