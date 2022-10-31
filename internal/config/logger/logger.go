package logger

import "github.com/sirupsen/logrus"

var l *logrus.Logger

func init() {
	l = logrus.New()
}

func GetLogger() *logrus.Logger {
	return l
}
