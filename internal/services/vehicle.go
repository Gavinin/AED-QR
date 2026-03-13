package services

type IVehicle interface {
	IsRunning() bool
	Open() error
	Lock() error
}
