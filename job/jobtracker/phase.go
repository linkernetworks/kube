package jobtracker

type Phase string

// Pod Phase defined in
// https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-phase
const PhasePending = Phase("pending")
const PhaseRunning = Phase("running")
const PhaseSucceeded = Phase("succeeded")
const PhaseFailed = Phase("failed")
const PhaseUnknown = Phase("unknown")
