package term

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/linkernetworks/logger"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"

	socketio "github.com/c9s/go-socket.io"

	corev1 "k8s.io/api/core/v1"
)

type SocketIoTermSession struct {
	Socket    socketio.Socket
	Stdout    *SocketIoWriter
	Stderr    *SocketIoWriter
	Stdin     *SocketIoReader
	SizeQueue *SocketIoSizeQueue
	TTY       bool
	Pod       *corev1.Pod
	Detached  bool
	CreatedAt time.Time
}

func NewSession(socket socketio.Socket, pod *corev1.Pod) *SocketIoTermSession {
	return &SocketIoTermSession{
		Socket:    socket,
		Stdin:     &SocketIoReader{Event: "term:stdin", Socket: socket, Buffer: make(chan []byte, 30)},
		Stdout:    &SocketIoWriter{Event: "term:stdout", Socket: socket, Buffer: bytes.NewBuffer([]byte{})},
		Stderr:    &SocketIoWriter{Event: "term:stderr", Socket: socket, Buffer: bytes.NewBuffer([]byte{})},
		SizeQueue: &SocketIoSizeQueue{C: make(chan *remotecommand.TerminalSize, 10)},
		TTY:       true,
		Pod:       pod,
		CreatedAt: time.Now(),
	}
}

func (s *SocketIoTermSession) NewExecutor(clientset *kubernetes.Clientset, restConfig *rest.Config, p ConnectRequestPayload) (remotecommand.Executor, error) {
	req := NewExecRequest(clientset, p)
	// logger.Debugln("Created exec request:", req.URL())
	return remotecommand.NewSPDYExecutor(restConfig, http.MethodPost, req.URL())
}

func (s *SocketIoTermSession) Detach() {
	s.Detached = true
}

func (s *SocketIoTermSession) Attach(socket socketio.Socket) {
	s.Detached = false
	s.Socket = socket
	s.Stdin.Socket = socket
	s.Stdout.Socket = socket
	s.Stderr.Socket = socket

	s.Socket.On("term:stdin", func(data string) {
		s.Stdin.Write([]byte(data))
	})

	s.Socket.On("term:resize", func(data string) {
		p := TermSizePayload{}
		err := json.Unmarshal([]byte(data), &p)
		if err != nil {
			logger.Errorf("term:resize error: %v", err)
			return
		}
		s.SizeQueue.Push(p.Columns, p.Rows)
	})
}

func (s *SocketIoTermSession) Terminate() {
	// ^U (21) and ^D (4)
	s.Stdin.Write([]byte{21, 4})
}
