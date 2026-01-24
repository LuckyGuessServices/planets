package router

import (
	"net/http"

	"github.com/LuckyGuessServices/planets/internal/luglog"
)

type handlerErrorLogger struct {
	err     error
	request *http.Request
	prefix  string
	extra   string
}

func newHandlerErrorLogger(err error, request *http.Request) *handlerErrorLogger {
	return &handlerErrorLogger{
		err:     err,
		request: request,
	}
}

func (logger *handlerErrorLogger) Prefix(prefix string) *handlerErrorLogger {
	logger.prefix = prefix + "; "

	return logger
}

func (logger *handlerErrorLogger) Extra(extraMessage string) *handlerErrorLogger {
	logger.extra = "\n" + extraMessage

	return logger
}

func (logger *handlerErrorLogger) Log() {
	luglog.Printf(
		"[API] %sURI '%s'; error --> %s%s",
		logger.prefix,
		logger.request.URL.Path,
		logger.err,
		logger.extra,
	)
}
