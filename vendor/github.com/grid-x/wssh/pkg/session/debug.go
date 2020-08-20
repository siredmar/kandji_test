package session

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
)

func (s *Session) spawnDebugDevice(listenAddr string, debugDeviceAddr string) {
	log := s.log.WithField("routine", "spawnDebugDevice")
	tunnelDevice, err := s.manager.DeviceTunnel(s.ID)

	if err != nil || tunnelDevice == nil {
		log.WithError(err).Error("could not get device tunnel")
		return
	}

	cmd := exec.Command(
		"./bin/device",
		fmt.Sprintf("-server-addr=%v", listenAddr),
		fmt.Sprintf("-device-addr=%v", debugDeviceAddr),
		fmt.Sprintf("-tID=%v", tunnelDevice.ID),
	)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.WithError(err).Error("create stderr pipe")
	}
	if err := cmd.Start(); err != nil {
		log.WithError(err).Error("start")
	}

	go func() {
		err := cmd.Wait()
		if err != nil {
			log.WithError(err).Error("exit")
		} else {
			log.Debug("exit")
		}
	}()

	b := bufio.NewReader(stderr)
	go func() {
		for {
			line, _, err := b.ReadLine()
			if err != nil {
				if err != io.EOF {
					log.WithError(err).Error("read")
				} else {
					log.Debug("EOF")
				}
				return
			}
			if len(line) > 0 {
				fmt.Println(string(line))
			}
		}
	}()
}
