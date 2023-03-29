package logger

import (
	"io"
	"os"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

var Logger = logrus.New()

func Init() {
	Logger.SetReportCaller(false)
	Logger.SetLevel(logrus.DebugLevel)

	Logger.SetFormatter(&nested.Formatter{
		HideKeys:        true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	file, _ := os.OpenFile("log.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	mw := io.MultiWriter(os.Stdout, file)
	Logger.SetOutput(mw)
}
