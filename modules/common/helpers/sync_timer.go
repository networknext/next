package helpers

import "time"

type SyncTimer struct{
	lastRun time.Time
	interval time.Duration
}

func NewSyncTimer(interval time.Duration) *SyncTimer{
	s := new(SyncTimer)
	s.lastRun = time.Now().Add(interval*5)
	s.interval = interval
	return s
}

func (s *SyncTimer)Run() {
	timeSince := time.Since(s.lastRun)
	if timeSince < s.interval && timeSince > 0{
		time.Sleep(s.interval - timeSince)
	}
	s.lastRun = time.Now()
}