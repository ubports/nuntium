package main

const (
	ErrorActivateContext = "x-ubports-nuntium-mms-error-activate-context"
	ErrorGetProxy        = "x-ubports-nuntium-mms-error-get-proxy"
	ErrorDownloadContent = "x-ubports-nuntium-mms-error-download-content"
	ErrorStorage         = "x-ubports-nuntium-mms-error-storage"
	ErrorForward         = "x-ubports-nuntium-mms-error-forward"
)

type standartizedError struct {
	error
	code string
}

func (e standartizedError) Code() string { return e.code }

type downloadError struct {
	standartizedError
}

func (e downloadError) AllowRedownload() bool { return true }
