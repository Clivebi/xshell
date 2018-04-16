xShell

xShell is a tool to open multi virtual terminal from ssh ,
so you can send command to all remote hosts only need one inputs.

xshell options [-h -i -b -w] hostslist
-h show help content
-i skip known_hosts check , or use you ssh knownhosts database(~/.ssh/known_hosts)
-b exit if any host can't connected
-w wait all hosts replay before execute next command
example use ssh ras key(in ~/.ssh/id_rsa) and (~/.ssh/known_hosts): 
xshell -b -w root@192.168.1.100:22 root@192.168.1.102:22
example user user and password (skip known_hosts check): 
xshell -i -b -w root@192.168.1.100:22@password root@192.168.1.102:22@password

ctrl+c will send to the remote hosts
if you want exit xShell,please input: xshell-exit

how to build

1 install golang :https://golang.org/

2 go get -u golang.org/x/crypto/ssh

3 go get -u github.com/chzyer/readline

4 go build -ldflags "-w -s" xshell.go


