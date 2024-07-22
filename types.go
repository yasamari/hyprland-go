package hyprland

import "net"

type RawRequest []byte

type RawResponse []byte

type RawData string

type EventType string

type ReceivedData struct {
	Type EventType
	Data RawData
}

// RequestClient is the main struct from hyprland-go.
// You may want to set 'Validate' as false to avoid (possibly costly)
// validations, at the expense of not reporting some errors in the IPC.
type RequestClient struct {
	Validate bool
	conn     *net.UnixAddr
}

// EventClient is the event struct from hyprland-go.
type EventClient struct {
	conn net.Conn
}

// Try to keep struct fields in the same order as the output for `hyprctl` for
// sanity.

type Client struct {
	Address        string        `json:"address"`
	Mapped         bool          `json:"mapped"`
	Hidden         bool          `json:"hidden"`
	At             []int         `json:"at"`
	Size           []int         `json:"size"`
	Workspace      WorkspaceType `json:"workspace"`
	Floating       bool          `json:"floating"`
	Pseudo         bool          `json:"pseudo"`
	Monitor        int           `json:"monitor"`
	Class          string        `json:"class"`
	Title          string        `json:"title"`
	InitialClass   string        `json:"initialClass"`
	InitialTitle   string        `json:"initialTitle"`
	Pid            int           `json:"pid"`
	Xwayland       bool          `json:"xwayland"`
	Pinned         bool          `json:"pinned"`
	Fullscreen     bool          `json:"fullscreen"`
	FullscreenMode int           `json:"fullscreenMode"`
	Grouped        []string      `json:"grouped"`
	Tags           []string      `json:"tags"`
	Swallowing     string        `json:"swallowing"`
	FocusHistoryId int           `json:"focusHistoryID"`
}

type CursorPos struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Monitor struct {
	Id               int           `json:"id"`
	Name             string        `json:"name"`
	Description      string        `json:"description"`
	Make             string        `json:"make"`
	Model            string        `json:"model"`
	Serial           string        `json:"serial"`
	Width            int           `json:"width"`
	Height           int           `json:"height"`
	RefreshRate      float64       `json:"refreshRate"`
	X                int           `json:"x"`
	Y                int           `json:"y"`
	ActiveWorkspace  WorkspaceType `json:"activeWorkspace"`
	SpecialWorkspace WorkspaceType `json:"specialWorkspace"`
	Reserved         []int         `json:"reserved"`
	Scale            float64       `json:"scale"`
	Transform        int           `json:"transform"`
	Focused          bool          `json:"focused"`
	DpmsStatus       bool          `json:"dpmsStatus"`
	Vrr              bool          `json:"vrr"`
	ActivelyTearing  bool          `json:"activelyTearing"`
	CurrentFormat    string        `json:"currentFormat"`
	AvailableModes   []string      `json:"availableModes"`
}

type Option struct {
	Option string `json:"option"`
	Int    int    `json:"int"`
	Set    bool   `json:"set"`
}

type Version struct {
	Branch        string   `json:"branch"`
	Commit        string   `json:"commit"`
	Dirty         bool     `json:"dirty"`
	CommitMessage string   `json:"commit_message"`
	CommitDate    string   `json:"commit_date"`
	Tag           string   `json:"tag"`
	Commits       string   `json:"commits"`
	Flags         []string `json:"flags"`
}

type Window struct {
	Client
}

type Workspace struct {
	WorkspaceType
	Monitor         string `json:"monitor"`
	MonitorID       int    `json:"monitorID"`
	Windows         int    `json:"windows"`
	HasFullScreen   bool   `json:"hasfullscreen"`
	LastWindow      string `json:"lastwindow"`
	LastWindowTitle string `json:"lastwindowtitle"`
}

type WorkspaceType struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}
