package xshell

import (
	"errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"io/ioutil"
)

type Shell struct {
	StdIn  *Console
	Stdout *Console
	host   *Host
	conf   *Config
}

var FAILED_NOTIFY = []byte("xshellconnecterror")

func NewShell(host *Host, conf *Config, stdout *Console) *Shell {
	shell := &Shell{
		StdIn:  NewConsole(),
		Stdout: stdout,
		host:   host,
		conf:   conf,
	}
	go shell.sshLoop()
	return shell
}

func (o *Shell) prepareAuthMethod() ([]ssh.AuthMethod, error) {
	if len(o.host.Password) != 0 {
		return []ssh.AuthMethod{ssh.Password(o.host.Password)}, nil
	}
	home, err := Home()
	if err != nil {
		return []ssh.AuthMethod{}, errors.New("unable get home path " + err.Error())
	}
	key, err := ioutil.ReadFile(home + "/.ssh/id_rsa")
	if err != nil {
		return []ssh.AuthMethod{}, errors.New("unable to read private key " + err.Error())
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return []ssh.AuthMethod{}, errors.New("unable to parse private key " + err.Error())
	}
	return []ssh.AuthMethod{ssh.PublicKeys(signer)}, nil
}

func (o *Shell) failed(err error) {
	t := o.host.Host + " " + err.Error() + "\n"
	o.Stdout.Write([]byte(t))
	o.Stdout.Write(FAILED_NOTIFY)
}

func (o *Shell) Write(p []byte) (n int, err error) {
	t := o.host.Host + " " + string(p)
	o.Stdout.Src <- []byte(t)
	return len(p), nil
}

func (o *Shell) closed() {
	o.Stdout.Write([]byte(o.host.Host + " connection closed"))

}

func (o *Shell) sshLoop() {

	callback := ssh.InsecureIgnoreHostKey()
	if !o.conf.IgnoreHostKey {
		home, err := Home()
		if err != nil {
			o.failed(errors.New("unable get home path " + err.Error()))
			return
		}
		callback, err = knownhosts.New(home + "/.ssh/known_hosts")
		if err != nil {
			o.failed(errors.New("unable read knownhosts " + err.Error()))
			return
		}
	}
	method, err := o.prepareAuthMethod()
	if err != nil {
		o.failed(err)
		return
	}
	config := &ssh.ClientConfig{
		User:            o.host.User,
		Auth:            method,
		HostKeyCallback: callback,
	}

	client, err := ssh.Dial("tcp", o.host.Host, config)
	if err != nil {
		o.failed(errors.New("unable to connect " + err.Error()))
		return
	}
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		o.failed(errors.New("unable start session " + err.Error()))
		return
	}
	defer session.Close()

	session.Stdout = o
	session.Stderr = o
	session.Stdin = o.StdIn

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	err = session.RequestPty("xterm", 100, 200, modes)
	if err != nil {
		o.failed(errors.New("unable request pty  " + err.Error()))
		return
	}

	err = session.Shell()
	if err != nil {
		o.failed(errors.New("unable start shell " + err.Error()))
		return
	}

	err = session.Wait()
	if err != nil {
		o.failed(errors.New("session return " + err.Error()))
	}
	o.closed()
}
