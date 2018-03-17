package term

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	corev1 "k8s.io/api/core/v1"

	scheme "k8s.io/client-go/kubernetes/scheme"
)

func NewExecRequest(clientset *kubernetes.Clientset, p ConnectRequestPayload) *rest.Request {
	rest := clientset.RESTClient()
	req := rest.Post().
		Prefix("/api/v1").
		Resource("pods").
		Name(p.PodName).
		Namespace("default").
		SubResource("exec").
		Param("container", p.ContainerName).
		Param("stdin", "true").
		Param("stdout", "true").
		Param("stderr", "true").
		Param("tty", "true").
		Param("command", "/bin/sh")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: p.ContainerName,
		Command:   []string{"sh"},

		// turn on the stdin if we have the input device connected
		Stdin: true,

		// read the stdout
		Stdout: true,

		// read the stderr
		Stderr: true,

		// tty is not allocated for the call
		TTY: true,
	}, scheme.ParameterCodec)
	return req
}
