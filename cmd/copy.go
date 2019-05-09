package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/grid-x/ds-api/pkg/ssh"
	"github.com/schollz/progressbar"
	"github.com/spf13/cobra"

	api "github.com/grid-x/gxctl/pkg/api"
	client "github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	template "github.com/grid-x/gxctl/pkg/template"
)

type Copy struct {
	Command *cobra.Command
}

const (
	ClientToDevice = iota
	DeviceToClient
)

func NewCopy(parent *cobra.Command, client *client.APIClient) *Copy {
	var copyCmd = &cobra.Command{
		Use:                   "copy ID SOURCE DESTINATION [OPTIONS]",
		DisableFlagsInUseLine: true,
		Short:                 "forwards files to devices",
		Example:               "# gxctl copy /tmp/test.conf c72:/opt/test.conf",
		Long:                  `TODO`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.MissingParameter("ID", "gxctl copy -h")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			copyCmdSource := args[0]
			copyCmdDestination := args[1]

			if strings.Contains(copyCmdSource, ":") && strings.Contains(copyCmdDestination, ":") {
				return errors.ServerError("You cannot specify the device in both source and destination parameters")
			}
			if !strings.Contains(copyCmdSource, ":") && !strings.Contains(copyCmdDestination, ":") {
				return errors.ServerError("Missing device id")
			}

			var id, source, dest string
			var direction int

			// eg. gxctl copy c72:/tmp/test.conf /opt/test.conf
			if strings.Contains(copyCmdSource, ":") {
				// Direction: Device -> Client
				direction = DeviceToClient
				splitted := strings.Split(copyCmdSource, ":")
				id = splitted[0]
				source = splitted[1]
			} else {
				source = copyCmdSource
			}

			// eg. gxctl copy /tmp/test.conf c72:/opt/test.conf
			if strings.Contains(copyCmdDestination, ":") {
				// Direction: Client -> Device
				direction = ClientToDevice
				splitted := strings.Split(copyCmdDestination, ":")
				id = splitted[0]
				dest = splitted[1]
			} else {
				dest = copyCmdDestination
			}

			//Lookup all existing devices to validate ids and autocomplete them if necessary
			devices, err := getDevices(client)
			if err != nil {
				return err
			}
			deviceIDs := devices.GetIds()

			deviceID, err := api.LookupID(id, deviceIDs)
			if err != nil {
				return err
			}

			if direction == ClientToDevice {
				if err := clientToDevice(source, dest, deviceID, client); err != nil {
					return err
				}
			}
			if direction == DeviceToClient {
				if err := deviceToClient(source, dest, deviceID, client); err != nil {
					return err
				}
			}

			fmt.Println("All done!")
			return nil
		},
	}

	copyCmd.SetHelpTemplate(template.HelpTemplate())
	copyCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(copyCmd)

	return &Copy{
		Command: copyCmd,
	}
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
		InitCommand:    fmt.Sprintf("/scp -S /dbclient root@127.0.0.1:%s /%s", source, tmpFileName),
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
		file, err := os.Create(dest)
		if err != nil {
			fmt.Println("Error while getting create file message", err)
			return
		}
		file.Close()

		sessionID := string(<-sessionChannelNew)

		// Issue inital create file message with reverse flag to let the agent start streaming
		msg := ssh.NewCreateFileMessage(sessionID, tmpFileName, true)
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

			var writeFile ssh.WriteToFileMessage

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

	tmpFileName := uuid.New().String()

	go func() {
		sessionID := string(<-sessionChannel)

		// Issue inital create file message
		msg := ssh.NewCreateFileMessage(sessionID, tmpFileName, false)
		m, err := json.Marshal(msg)
		if err != nil {
			fmt.Println("Unknown error: ", err)
			return
		}
		socketInputChannel <- m

		var size, written int64
		var progress float32
		var bar *progressbar.ProgressBar

		fi, err := srcFile.Stat()
		if err != nil {
			size = 0
			fmt.Println("Could not determine file size to display progress")
		} else {
			bar = progressbar.New(100)
			size = fi.Size()
		}

		// Using a 32*1024 bytes buffer as done in io.Copy()
		buf := make([]byte, 32*1024)
		for {
			n, err := srcFile.Read(buf)
			if err != nil {
				// EOF reached
				msg := ssh.NewWriteToFileMessage(sessionID, tmpFileName, nil, true)
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
					progress = float32(written) / float32(size) * 100
					bar.Set64(int64(progress))
				}
				msg := ssh.NewWriteToFileMessage(sessionID, tmpFileName, buf[0:n], false)
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
		InitCommand:    "",
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
		InitCommand:    fmt.Sprintf("/scp -S /dbclient /%s root@127.0.0.1:%s", tmpFileName, dest),
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
