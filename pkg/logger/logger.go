package logger

import (
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/junwei0117/logs-collector/pkg/configs"
	"github.com/sirupsen/logrus"
)

var Logger = logrus.New()

func Init() {
	Logger.SetReportCaller(configs.ReportCaller)
	if configs.Debug {
		Logger.SetLevel(logrus.DebugLevel)
	} else {
		Logger.SetLevel(logrus.InfoLevel)
	}

	Logger.SetFormatter(&nested.Formatter{
		HideKeys:        true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
}
