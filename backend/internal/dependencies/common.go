package dependencies

import (
	"medbratishka/pkg/logger"
	"medbratishka/pkg/time_manager"
)

func (d *Dependencies) Logger() logger.Logger {
	return d.logger
}

func (d *Dependencies) TimeManager() time_manager.TimeManager {
	if d.timeManager == nil {
		d.timeManager = time_manager.New(3)
	}
	return d.timeManager
}
