package ports

type Notifier interface {
	SendToDevice(deviceID string, msg []byte) error
	Broadcast(msg []byte)
	IsConnected(deviceID string) bool
}
