package term

import (
	"io"

	"bitbucket.org/linkernetworks/aurora/src/logger"

	socketio "github.com/c9s/go-socket.io"
)

type SocketIoReader struct {
	Event  string
	Socket socketio.Socket
	Buffer chan []byte
}

func (r *SocketIoReader) Write(p []byte) (n int, err error) {
	r.Buffer <- p
	return len(p), nil
}

func (r *SocketIoReader) Read(p []byte) (n int, err error) {
	data := <-r.Buffer
	n = copy(p, data)
	return n, err
}

type SocketIoWriter struct {
	Event  string
	Socket socketio.Socket
}

func (w *SocketIoWriter) Write(p []byte) (n int, err error) {
	data := string(p)
	if err := w.Socket.Emit(w.Event, data); err != nil {
		if err != io.EOF {
			logger.Errorf("term writer emit error: %v", err)
		}
		return 0, err
	}
	return len(p), err
}
