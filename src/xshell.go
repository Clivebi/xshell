package xshell

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/signal"
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
				fmt.Println("have some error exit...")
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

func (o *XShell) CommandLoop() {
	r := bufio.NewReader(os.Stdin)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		for {
			<-c
			o.writeCommand([]byte{0x3})
		}
	}()
	for {
		l, err := r.ReadString('\n')
		if err != nil {
			break
		}
		if strings.ToLower(l) == "xshell-exit\n" {
			fmt.Println("~Bye~")
			break
		}
		o.writeCommand([]byte(l))
	}
}

func (o *XShell) writeCommand(cmd []byte) {
	if o.conf.WaitForReplay && !o.checkEcho() {
		fmt.Println("please wait for all host replay message received!")
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
