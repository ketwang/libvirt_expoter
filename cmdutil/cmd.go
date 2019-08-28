package cmdutil

import (
	"bytes"
	"os/exec"
	"time"
)

func Command(cmd string, args []string) ([]byte, []byte, error) {
	var stdOutput bytes.Buffer
	var errOutput bytes.Buffer

	metricCmdExecTotal.WithLabelValues(cmd, args[0]).Inc()

	start := time.Now()

	d := exec.Command(cmd, args...)
	d.Stdout = &stdOutput
	d.Stderr = &errOutput
	err := d.Run()

	metricCmdExecLatency.WithLabelValues(cmd, args[0]).Observe(time.Now().Sub(start).Seconds())

	return stdOutput.Bytes(), errOutput.Bytes(), err
}
