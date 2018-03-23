package xshell

import (
	"bytes"
	"fmt"
	"github.com/chzyer/readline"
	"io"
	"os"
	"strings"
	"sync"
)

type XShell struct {
	stdOut *Console
	shells []*Shell
	conf   *Config
	lock   *sync.RWMutex
	echos  map[string]int
}

func (o *XShell) outputLoop() {
	for {
		buf := <-o.stdOut.Src
		if buf == nil {
			continue
		}
		if bytes.Compare(buf, FAILED_NOTIFY) == 0 {
			if o.conf.BreakOnFailed {
				fmt.Println("An error occurred,exit ...")
				os.Exit(-1)
				break
			}
		}
		msg := string(buf)
		if o.conf.WaitForReplay {
			o.parseEcho(msg)
		}
		if strings.Contains(msg, "$") {
			fmt.Println(string(buf))
		} else {
			fmt.Print(string(buf))
		}
	}
}

func (o *XShell) parseEcho(msg string) {
	ls := strings.Split(msg, " ")
	if len(ls) < 1 {
		return
	}
	host := ""
	for _, v := range o.shells {
		if v.host.Host == ls[0] {
			host = ls[0]
			break
		}
	}
	if len(host) == 0 {
		return
	}
	o.lock.Lock()
	defer o.lock.Unlock()
	v, _ := o.echos[host]
	o.echos[host] = v + 1
}

func (o *XShell) checkEcho() bool {
	o.lock.Lock()
	defer o.lock.Unlock()
	if len(o.echos) != len(o.shells) {
		fmt.Println("please wait for all host replay message received!")
		for _, v := range o.shells {
			_, ok := o.echos[v.host.Host]
			if !ok {
				fmt.Println(v.host.Host, " replay message not received!")
			}
		}
		return false
	} else {
		dw := false
		dv := 0
		for _, v := range o.echos {
			if dv == 0 {
				dv = v
			} else {
				if dv != v {
					dw = true
				}
			}
		}
		if dw {
			fmt.Println("warning : replay message count not equal ")
		}
		return true
	}
}

func (o *XShell) clearEcho() {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.echos = make(map[string]int)
}

func (o *XShell) CommandLoop() error {
	path, _ := Home()
	path += "/.xshell_history"
	l, err := readline.NewEx(&readline.Config{
		Prompt:            "",
		HistoryFile:       path,
		InterruptPrompt:   "ctrl+C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	})
	if err != nil {
		return err
	}
	defer l.Close()
	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				o.writeCommand([]byte{0x3})
			}
			continue
		} else if err == io.EOF {
			continue
		}
		if len(line) == 0 || line[len(line)-1] != '\n' {
			line += "\n"
		}
		if strings.ToLower(line) == "xshell-exit\n" {
			fmt.Println("~Bye~")
			break
		}
		o.writeCommand([]byte(line))
	}
	return nil
}

func (o *XShell) writeCommand(cmd []byte) {
	if o.conf.WaitForReplay && !o.checkEcho() {
		return
	}
	o.clearEcho()
	for _, v := range o.shells {
		if v.StdIn != nil {
			v.StdIn.Write(cmd)
		}
	}
}

func NewXShell(Hosts []*Host, conf *Config) *XShell {
	ret := &XShell{
		stdOut: NewConsole(),
		shells: make([]*Shell, len(Hosts)),
		conf:   conf,
		echos:  make(map[string]int),
		lock:   &sync.RWMutex{},
	}
	for i, v := range Hosts {
		ret.shells[i] = NewShell(v, conf, ret.stdOut)
	}
	go ret.outputLoop()
	return ret
}
