
SSH proxy to serial console of OpenNebula machines. The server runs on port 2222 and listens for ssh connections. The username is the ID of the VM to connect to its serial console. The ssh keys in the context section are used for authentication. To connect to the serial console it start an ssh connection to the host where the VM is running and executes:

    virsh -c qemu:///system console one-<id>

## Requirements

* Create an rsa key without password in the same directory where you start the server

```
$ ssh-keygen -f ./id_rsa
```

* The user running the server should have OpenNebula credentials to retrieve VM info and ssh configured to connect to the hosts. Tested with `oneadmin` user

## Binary

https://downloads.zooloo.org/one-serial.xz

## VM preparation

The VM template should have serial console enabled. Add this to the template:

```
RAW=[
  DATA="<devices><serial type='pty'><target port='0'/></serial><console type='pty'><target type='serial' port='0'/></console></devices>",
  TYPE="kvm" ]
```

The OS should be prepared to use serial console. To make it work in the CentOS 7 image provided in the OpenNebula marketplace you can execute this command:

```
systemctl enable serial-getty@ttyS0.service --now
```

## Running

Start the server:

```
$ ./main
```

Connect to VM 654:

```
ssh -p 654@<your frontend>
```

## Acknowledgments

The code is heavily based (copied) from these examples:

 * https://github.com/gliderlabs/ssh/blob/master/_examples/ssh-pty/pty.go
 * https://github.com/golang/crypto/blob/master/ssh/example_test.go

Libraries used:

 * https://github.com/OpenNebula/goca
 * https://github.com/gliderlabs/ssh
 * https://github.com/kr/pty


