package hyprland

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

const (
	BUF_SIZE     = 8192
	MAX_COMMANDS = 30
)

// IPCClient is the main struct from hyprland-go.
// You may want to set 'Validate' as false to avoid (possibly costly)
// validations, at the expense of not reporting some errors in the IPC.
type IPCClient struct {
	Validate    bool
	requestConn *net.UnixAddr
	eventConn   net.Conn
}

func must1[T any](v T, err error) T {
	must(err)
	return v
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func prepareRequests(command string, params []string) (requests [][]byte) {
	if command == "" {
		panic("empty command")
	}
	switch len(params) {
	case 0:
		requests = append(requests, []byte(command))
	case 1:
		requests = append(requests, []byte(fmt.Sprintf("%s %s", command, params[0])))
	default:
		// Hyprland IPC has a hidden limit for commands, so we are
		// splitting the commands in multiple requests if the user pass
		// more commands that it is supported
		var buffer bytes.Buffer
		for i := 0; i < len(params); i += MAX_COMMANDS {
			end := i + MAX_COMMANDS
			if end > len(params) {
				end = len(params)
			}

			buffer.Reset()
			buffer.WriteString("[[BATCH]]")
			for j := i; j < end; j++ {
				buffer.WriteString(command)
				buffer.WriteString(" ")
				buffer.WriteString(params[j])
				buffer.WriteString(";")
			}

			requests = append(requests, buffer.Bytes())
		}
	}
	return requests
}

func (c *IPCClient) validateResponse(params []string, response []byte) error {
	if !c.Validate {
		return nil
	}

	// Empty response
	if len(response) == 0 {
		return errors.New("empty response")
	}
	// Count the number of "ok" we got in response
	got := strings.Count(string(response), "ok")
	want := len(params)
	// Commands without parameters still have a "ok" response
	if want == 0 {
		want = 1
	}
	// If we had less than expected number of "ok" results, it means
	// something went wrong
	if got < want {
		return errors.New(
			fmt.Sprintf(
				"got ok: %d, want: %d, response: %s",
				got,
				want,
				response,
			),
		)
	}
	return nil
}

func unmarshalResponse(response []byte, v any) (err error) {
	if len(response) == 0 {
		return errors.New("empty response")
	}

	err = json.Unmarshal(response, &v)
	if err != nil {
		return fmt.Errorf("error during unmarshal: %w", err)
	}
	return nil
}

func (c *IPCClient) doRequest(command string, params ...string) (response []byte, err error) {
	requests := prepareRequests(command, params)
	for _, req := range requests {
		resp, err := c.Request(req)
		if err != nil {
			return nil, fmt.Errorf("error while doing request: %w", err)
		}
		response = append(response, resp...)
	}
	return response, nil
}

// Initiate a new client or panic.
// This should be the preferred method for user scripts, since it will
// automatically find the proper socket to connect and use the
// HYPRLAND_INSTANCE_SIGNATURE for the current user.
// If you need to connect to arbitrary user instances or need a method that
// will not panic on error, use [NewClient] instead.
func MustClient() *IPCClient {
	his := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	if his == "" {
		panic("HYPRLAND_INSTANCE_SIGNATURE is empty, are you using Hyprland?")
	}

	// https://github.com/hyprwm/Hyprland/blob/83a5395eaa99fecef777827fff1de486c06b6180/hyprctl/main.cpp#L53-L62
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		user := must1(user.Current()).Uid
		runtimeDir = filepath.Join("/run/user", user)
	}

	return must1(
		NewClient(
			filepath.Join(runtimeDir, "hypr", his, ".socket.sock"),
			filepath.Join(runtimeDir, "hypr", his, ".socket2.sock"),
		),
	)
}

// Initiate a new client.
// Receive as parameters a requestSocket that is generally localised in
// '$XDG_RUNTIME_DIR/hypr/$HYPRLAND_INSTANCE_SIGNATURE/.socket.sock' and
// eventSocket that is generally localised in
// '$XDG_RUNTIME_DIR/hypr/$HYPRLAND_INSTANCE_SIGNATURE/.socket2.sock'.
func NewClient(requestSocket, eventSocket string) (*IPCClient, error) {
	if requestSocket == "" || eventSocket == "" {
		return nil, errors.New("empty request or event socket")
	}

	conn, err := net.Dial("unix", eventSocket)
	if err != nil {
		return nil, fmt.Errorf("error while connecting to socket: %w", err)
	}

	return &IPCClient{
		Validate: true,
		requestConn: &net.UnixAddr{
			Net:  "unix",
			Name: requestSocket,
		},
		eventConn: conn,
	}, nil
}

// Low-level request method, should be avoided unless there is no alternative.
// Receives a byte array as parameter that should be a valid command similar to
// 'hyprctl' command, e.g.: 'hyprctl dispatch exec kitty' will be
// '[]byte("dispatch exec kitty")'.
// Keep in mind that there is no validation. In case of an invalid request, the
// response will generally be something different from "ok".
func (c *IPCClient) Request(request []byte) (response []byte, err error) {
	if len(request) == 0 {
		return nil, errors.New("empty request")
	}

	// Connect to the request socket
	conn, err := net.DialUnix("unix", nil, c.requestConn)
	defer conn.Close()
	if err != nil {
		return nil, fmt.Errorf("error while connecting to socket: %w", err)
	}

	// Send the request to the socket
	request = append([]byte{'j', '/'}, request...)
	_, err = conn.Write(request)
	if err != nil {
		return nil, fmt.Errorf("error while writing to socket: %w", err)
	}

	// Get the response back
	var resp bytes.Buffer
	buf := make([]byte, BUF_SIZE)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		resp.Write(buf[:n])
		if n < BUF_SIZE {
			break
		}
	}

	return resp.Bytes(), nil
}

// Get option command, similar to 'hyprctl activewindow'.
// Returns a [Window] object.
func (c *IPCClient) ActiveWindow() (w Window, err error) {
	response, err := c.doRequest("activewindow")
	if err != nil {
		return w, err
	}
	return w, unmarshalResponse(response, &w)
}

// Get option command, similar to 'hyprctl activeworkspace'.
// Returns a [Workspace] object.
func (c *IPCClient) ActiveWorkspace() (w Workspace, err error) {
	response, err := c.doRequest("activeworkspace")
	if err != nil {
		return w, err
	}
	return w, unmarshalResponse(response, &w)
}

// Get option command, similar to 'hyprctl clients'.
// Returns a [Client] object.
func (c *IPCClient) Clients() (cl []Client, err error) {
	response, err := c.doRequest("clients")
	if err != nil {
		return cl, err
	}
	return cl, unmarshalResponse(response, &cl)
}

// Get option command, similar to 'hyprctl cursorpos'.
// Returns a [CursorPos] object.
func (c *IPCClient) CursorPos() (cu []CursorPos, err error) {
	response, err := c.doRequest("cursorpos")
	if err != nil {
		return cu, err
	}
	return cu, unmarshalResponse(response, &cu)
}

// Dispatch commands, similar to 'hyprctl dispatch'.
// Accept multiple commands at the same time, in this case it will use batch
// mode, similar to 'hyprctl dispatch --batch'.
func (c *IPCClient) Dispatch(params ...string) error {
	response, err := c.doRequest("dispatch", params...)
	if err != nil {
		return err
	}
	return c.validateResponse(params, response)
}

// Get option command, similar to 'hyprctl getoption'.
// Returns an [Option] object.
func (c *IPCClient) GetOption(name string) (o Option, err error) {
	response, err := c.doRequest("getoption", name)
	if err != nil {
		return o, err
	}
	return o, unmarshalResponse(response, &o)
}

// Kill command, similar to 'hyprctl kill'.
// Will NOT wait for the user to click in the window.
func (c *IPCClient) Kill() error {
	response, err := c.doRequest("kill")
	if err != nil {
		return err
	}
	return c.validateResponse(nil, response)
}

// Reload command, similar to 'hyprctl reload'.
func (c *IPCClient) Reload() error {
	response, err := c.doRequest("reload")
	if err != nil {
		return err
	}
	return c.validateResponse(nil, response)
}

// Get option command, similar to 'hyprctl version'.
// Returns an [Version] object.
func (c *IPCClient) Version() (v Version, err error) {
	response, err := c.doRequest("version")
	if err != nil {
		return v, err
	}
	return v, unmarshalResponse(response, &v)
}

// Get option command, similar to 'hyprctl splash'.
func (c *IPCClient) Splash() (s string, err error) {
	response, err := c.doRequest("splash")
	if err != nil {
		return "", err
	}
	return string(response), nil
}
