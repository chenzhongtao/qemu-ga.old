package main

import (
	"encoding/json"
	"io"
	"log"
	"os"

	flags "github.com/jessevdk/go-flags"
)

var options struct {
	Verbose    bool     `short:"v" long:"verbose" description:"log extra debugging information"`
	Version    bool     `short:"V" long:"version" description:"print version information and exit"`
	Help       bool     `short:"h" long:"help" description:"display this help and exit"`
	Daemon     bool     `short:"d" long:"daemonize" description:"become a daemon"`
	Blacklist  []string `short:"b" long:"blacklist" description:"comma-separated list of RPCs to disable (no spaces, \"?\" to list available RPCs)"`
	StateDir   string   `short:"t" long:"statedir" default:"/var/run" description:"specify dir to store state information (absolute paths only, default is /var/run)"`
	FreezeHook string   `short:"F" long:"fsfreeze-hook" default:"/etc/qemu/fsfreeze-hook" description:"enable fsfreeze hook. Accepts an optional argument that specifies script to run on freeze/thaw. Script will be called with 'freeze'/'thaw' arguments accordingly. (default is /etc/qemu/fsfreeze-hook) If using -F with an argument, do not follow -F with a space. (for example: -F/var/run/fsfreezehook.sh)"`
	Method     string   `short:"m" long:"method" default:"virtio-serial" description:"transport method: one of unix-listen, virtio-serial, or isa-serial (virtio-serial is the default)"`
	Path       string   `short:"p" long:"path" default:"/dev/virtio-ports/org.qemu.guest_agent.0" description:"device/socket path (the default for virtio-serial is: /dev/virtio-ports/org.qemu.guest_agent.0, the default for isa-serial is: /dev/ttyS0)"`
	LogFile    string   `short:"l" long:"logfile" default:"stderr" description:"set logfile path, logs to stderr by default"`
	PidFile    string   `short:"f" long:"pidfile" default:"/var/run/qemu-ga.pid" description:"specify pidfile (default is /var/run/qemu-ga.pid)"`
}

type Request struct {
	Execute   string                 `json:"execute"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type Response struct {
	Return interface{} `json:"return"`
}

func main() {
	parser := flags.NewParser(&options, flags.PrintErrors)
	_, err := parser.Parse()
	if err != nil {
		log.Printf(err.Error())
		os.Exit(1)
	}

	if options.Help {
		parser.WriteHelp(os.Stdout)
		os.Exit(0)
	}
	f, err := os.OpenFile(options.Path, os.O_RDWR, os.ModeCharDevice|0600)
	if err != nil {
		log.Printf(err.Error())
		os.Exit(1)
	}

	dec := json.NewDecoder(f)
	dec.UseNumber()
	//	w := io.MultiWriter(f, os.Stdout)
	enc := json.NewEncoder(f)

	defer f.Close()
	var req Request

	for {
		if err := dec.Decode(&req); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		for _, cmd := range commands {
			if cmd.Name == req.Execute && cmd.Func != nil {
				enc.Encode(cmd.Func(req.Arguments))
			}
		}
	}
}