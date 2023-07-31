package service

type ProxyServiceInterface interface {
	Run() error
	Stop() error
}
