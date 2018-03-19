package main

import (
	"fmt"
	"github.com/Clivebi/xshell/src"
	"os"
)

func showHelp() {
	fmt.Println("xshell options [-h -i -b -w] hostslist")
	fmt.Println("-h show help content")
	fmt.Println("-i skip known_hosts check , or use you ssh knownhosts database(~/.ssh/known_hosts)")
	fmt.Println("-b exit if any host can't connected")
	fmt.Println("-w wait all hosts replay before execute next command")
	fmt.Println("example use ssh ras key(in ~/.ssh/id_rsa):")
	fmt.Println("xshell -i -b -w root@192.168.1.100:22 root@192.168.1.102:22")
	fmt.Println("example user user and password: ")
	fmt.Println("xshell -i -b -w root@192.168.1.100:22@password root@192.168.1.102:22@password")
}
func main() {
	hosts := []*xshell.Host{}
	conf := &xshell.Config{}
	if len(os.Args) == 1 {
		showHelp()
		return
	}
	for i, v := range os.Args {
		if i < 1 {
			continue
		}
		switch v {
		case "-h":
			showHelp()
			return
		case "-i":
			conf.IgnoreHostKey = true
		case "-b":
			conf.BreakOnFailed = true
		case "-w":
			conf.WaitForReplay = true
		default:
			o, err := xshell.NewHost(v)
			if err != nil {
				fmt.Println(err)
				showHelp()
				return
			}
			hosts = append(hosts, o)
		}
	}
	if len(hosts) == 0 {
		showHelp()
		return
	}
	mshell := xshell.NewXShell(hosts, conf)
	mshell.CommandLoop()
}
