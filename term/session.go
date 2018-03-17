package term

import (
	"encoding/json"
	"net/http"

	"bitbucket.org/linkernetworks/aurora/src/logger"

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
}

func NewSession(socket socketio.Socket, pod *corev1.Pod) *SocketIoTermSession {
	return &SocketIoTermSession{
		Socket:    socket,
		Stdin:     &SocketIoReader{Event: "term:stdin", Socket: socket, Buffer: make(chan []byte, 30)},
		Stdout:    &SocketIoWriter{Event: "term:stdout", Socket: socket},
		Stderr:    &SocketIoWriter{Event: "term:stderr", Socket: socket},
		SizeQueue: &SocketIoSizeQueue{C: make(chan *remotecommand.TerminalSize, 10)},
		TTY:       true,
		Pod:       pod,
	}
}

func (s *SocketIoTermSession) Connect(clientset *kubernetes.Clientset, restConfig *rest.Config, p ConnectRequestPayload) error {
	req := NewExecRequest(clientset, p)

	logger.Infoln("Created request:", req.URL())

	exec, err := remotecommand.NewSPDYExecutor(restConfig, http.MethodPost, req.URL())
	if err != nil {
		return err
	}

	s.Socket.Emit("term:connected")

	return exec.Stream(remotecommand.StreamOptions{
		Stdin:             s.Stdin,
		Stdout:            s.Stdout,
		Stderr:            s.Stderr,
		Tty:               s.TTY,
		TerminalSizeQueue: s.SizeQueue,
	})
}

func (s *SocketIoTermSession) attach() {
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
