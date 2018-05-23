package term

type Payload struct {
	Data string `json:"data"`
}

type ConnectRequestPayload struct {
	Namespace     string `json:"namespace"`
	PodName       string `json:"pod"`
	ContainerName string `json:"container"`
	Command       string `json:"command,omitempty"`
	SessionID     string `json:"session,omitempty"`
}

type TermSizePayload struct {
	Columns uint16 `json:"cols"`
	Rows    uint16 `json:"rows"`
}
