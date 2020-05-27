package action

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	api "github.com/grid-x/gxctl/pkg/api"
	client "github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/service"
)

const (
	ClientToDevice = iota
	DeviceToClient
)

func Copy(s *service.Service, sourceID string, destID string) error {
	var id, source, dest string
	var direction int

	// eg. gxctl copy c72:/tmp/test.conf /opt/test.conf
	if strings.Contains(sourceID, ":") {
		// Direction: Device -> Client
		direction = DeviceToClient
		splitted := strings.Split(sourceID, ":")
		id = splitted[0]
		source = splitted[1]
	} else {
		source = sourceID
	}

	// eg. gxctl copy /tmp/test.conf c72:/opt/test.conf
	if strings.Contains(destID, ":") {
		// Direction: Client -> Device
		direction = ClientToDevice
		splitted := strings.Split(destID, ":")
		id = splitted[0]
		dest = splitted[1]
	} else {
		dest = destID
	}

	//Lookup all existing devices to validate ids and autocomplete them if necessary
	devices, err := getDevices(s.Client)
	if err != nil {
		return err
	}
	deviceIDs := devices.GetIds()

	deviceID, err := api.LookupID(id, deviceIDs)
	if err != nil {
		return err
	}

	if direction == ClientToDevice {
		if err := clientToDevice(source, dest, deviceID, s.Client); err != nil {
			return err
		}
	}
	if direction == DeviceToClient {
		if err := deviceToClient(source, dest, deviceID, s.Client); err != nil {
			return err
		}
	}

	fmt.Println("All done!")
	return nil
}

func deviceToClient(source, dest, deviceID string, client *client.APIClient) error {
	socketInputChannel := make(chan []byte)
	socketOutputChannel := make(chan []byte)
	sessionChannel := make(chan []byte)
	tmpFileName := uuid.New().String()

	// First copy the source file from the devie to within the container... Time for some SCP magic to happen
	conf := &SSHConfig{
		Client:         client,
		DeviceID:       deviceID,
		WaitText:       "Connecting to the device...",
		ReadyText:      "Starting file transfer...",
		InitCommand:    fmt.Sprintf("/scp -P 22222 -S /dbclient root@127.0.0.1:%s /%s", source, tmpFileName),
		InputChannel:   socketInputChannel,
		OutputChannel:  socketOutputChannel,
		SessionChannel: sessionChannel,
		Silent:         false,
	}

	go func() {
		// Unblock sessionChannel for SCP Session
		<-sessionChannel
	}()

	// Ignore output messages but make sure channel is not blocked
	go func() {
		for {
			_, ok := <-socketOutputChannel
			if !ok {
				break
			}
		}
	}()

	if err := createSession(conf); err != nil {
		return err
	}

	// At this point the file will be stored within the container. Reinit channels and stream it to the client
	socketInputChannelNew := make(chan []byte)
	socketOutputChannelNew := make(chan []byte)
	sessionChannelNew := make(chan []byte)

	go func() {
		// Create file
		if stat, err := os.Stat(dest); err == nil && stat.IsDir() {
			dest = filepath.Join(dest, filepath.Base(source))
		}
		file, err := os.Create(dest)
		if err != nil {
			fmt.Println("Error while getting create file message", err)
			return
		}
		file.Close()

		sessionID := string(<-sessionChannelNew)

		// Issue inital create file message with reverse flag to let the agent start streaming
		msg := api.NewCreateFileMessage(sessionID, tmpFileName, true, 0)
		m, err := json.Marshal(msg)
		if err != nil {
			fmt.Println("Unknown error: ", err)
			return
		}
		socketInputChannelNew <- m

		f, err := os.OpenFile(dest, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("Unknown error: ", err)
			return
		}
		defer f.Close()

		for {
			message := <-socketOutputChannelNew

			var writeFile api.WriteToFileMessage

			if err := json.Unmarshal(message, &writeFile); err == nil {
				// If err != nil we might have got an plain process output which we are going to ignore
				if writeFile.EOF {
					break
				}

				if _, err := f.Write(writeFile.Content); err != nil {
					fmt.Println("Error while writing to file", err)
					break
				}
			}
		}
		fmt.Println("All done")
		os.Exit(0)
	}()

	conf = &SSHConfig{
		Client:         client,
		DeviceID:       deviceID,
		WaitText:       "Connecting to the device...",
		InitCommand:    "",
		InputChannel:   socketInputChannelNew,
		OutputChannel:  socketOutputChannelNew,
		SessionChannel: sessionChannelNew,
		Silent:         true,
	}

	if err := createSession(conf); err != nil {
		return err
	}

	return nil
}

func clientToDevice(source, dest, deviceID string, client *client.APIClient) error {
	srcFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	socketInputChannel := make(chan []byte)
	socketOutputChannel := make(chan []byte)
	sessionChannel := make(chan []byte)

	// Ignore output messages but make sure channel is not blocked
	go func() {
		for {
			<-socketOutputChannel
		}
	}()

	tmpFolder := uuid.New().String()
	fileName := filepath.Base(srcFile.Name())

	targetPath := tmpFolder + "/" + fileName

	go func() {
		sessionID := string(<-sessionChannel)

		// Issue inital create file message
		msg := api.NewCreateFileMessage(sessionID, targetPath, false, 0)
		m, err := json.Marshal(msg)
		if err != nil {
			fmt.Println("Unknown error: ", err)
			return
		}
		socketInputChannel <- m

		var size, written int64

		fi, err := srcFile.Stat()
		if err != nil {
			size = 0
			fmt.Println("Could not determine file size to display progress")
		} else {
			size = fi.Size()
		}

		// Using a 32*1024 bytes buffer as done in io.Copy()
		buf := make([]byte, 32*1024)
		for {
			n, err := srcFile.Read(buf)
			if err != nil {
				// EOF reached
				msg := api.NewWriteToFileMessage(sessionID, targetPath, nil, true)
				m, err := json.Marshal(msg)
				if err != nil {
					fmt.Println("Unknown error: ", err)
					break
				}
				socketInputChannel <- m
				break
			}
			if n > 0 {
				if size != 0 {
					written += int64(n)
				}
				msg := api.NewWriteToFileMessage(sessionID, targetPath, buf[0:n], false)
				m, err := json.Marshal(msg)
				if err != nil {
					fmt.Println("Unknown error: ", err)
					break
				}
				socketInputChannel <- m
			}
		}
		fmt.Println("\r\nThe file has been successfully transferred. Finishing...")
	}()

	conf := &SSHConfig{
		Client:         client,
		DeviceID:       deviceID,
		WaitText:       "Connecting to the device...",
		ReadyText:      "Starting file transfer...",
		InitCommand:    "/dbclient -p 22222 -y root@127.0.0.1",
		InputChannel:   socketInputChannel,
		OutputChannel:  socketOutputChannel,
		SessionChannel: sessionChannel,
		Silent:         false,
	}

	if err := createSession(conf); err != nil {
		return err
	}

	// At this point we've copied the file into the container... Time for some SCP magic to happen
	conf = &SSHConfig{
		Client:         client,
		DeviceID:       deviceID,
		WaitText:       "Connecting to the device...",
		InitCommand:    fmt.Sprintf("/scp -P 22222 -S /dbclient /%s root@127.0.0.1:%s", targetPath, dest),
		InputChannel:   socketInputChannel,
		OutputChannel:  socketOutputChannel,
		SessionChannel: sessionChannel,
		Silent:         true,
	}

	go func() {
		// Unblock sessionChannel for SCP Session
		<-sessionChannel
	}()

	if err := createSession(conf); err != nil {
		return err
	}

	return nil
}
