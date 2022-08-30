package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

var login_flag = false
var current_id string

var debug bool = false

func main() {
	if debug {
		login_flag = true
		current_id = "debug"
	}

	conn := connectServer("127.0.0.1:5000")
	defer conn.Close()
	for true {
		printPrompt(login_flag, current_id)
		err := sendCommand(getInput(), conn)
		if err != nil {
			//if err.Error() == "empty_send" {
			//	continue
			//}
			// 这里如果出现不是我定义的err应该会直接panic，因为可能没有Error()方法
			// 算了直接一刀切全部continue就完了
			continue
		}

		data_str, cmd_slice, cmd_check, err := recvData(conn)

		if err != nil {
			fmt.Print(err)
		}
		if cmd_check == false {
			fmt.Println(data_str)
		} else {
			processCmd(cmd_slice)
		}

	}

}
func connectServer(remote_addr string) *net.TCPConn {
	remote_tcpaddr, err := net.ResolveTCPAddr("tcp4", remote_addr)
	if err != nil {
		return nil
	}
	conn, err := net.DialTCP("tcp4", nil, remote_tcpaddr)
	if err != nil {
		return nil
	} else {
		return conn
	}

}

func printPrompt(login_flag bool, id string) {
	if login_flag {
		fmt.Print("@[" + id + "]>>> ")
	} else {
		fmt.Print("@[?]>>> ")
	}
}

func getInput() (input_str string) {
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	input_str = input.Text()
	return
}

func sendCommand(cmd string, conn *net.TCPConn) (err error) {
	// 需要作判断了，如果以及登录，那么send的指令应该带上 -id 参数
	appendix := ""
	// 复用server里面processCmdStrToSlice的代码
	cmd_slice := make([]string, 0)
	tmp := bufio.NewScanner(strings.NewReader(cmd))
	tmp.Split(bufio.ScanWords)
	for tmp.Scan() {
		cmd_slice = append(cmd_slice, tmp.Text())
	}

	if len(cmd_slice) != 0 {
		switch cmd_slice[0] {
		// 只有空格导致越界
		// 对于某些cmd需要添加后缀信息或做些其他的事情
		case "sendmsg":
			fallthrough
		case "checkmsg":
			fallthrough
		case "whoami":
			if login_flag {
				appendix = " -id " + current_id
			} else {
				fmt.Println("Loooooogin required, wanna join?")
				return errors.New("empty_send")
			}
		default: // 不需要添加的统统不写，进入default分支
			appendix = ""
		}
	}

	length, err := conn.Write([]byte(cmd + appendix))
	if err != nil {
		return err
	} else if length == 0 {
		return errors.New("empty_send")
	} else {
		return nil
	}
}

func recvData(conn net.Conn) (data string, may_cmd_slice []string, command bool, err error) { // 我直接复用getCmdString
	buff_size := 128
	var data_buff []byte
	for true { // 缓冲区
		buff := make([]byte, buff_size)
		length, err := conn.Read(buff)
		if err != nil {
			return "", nil, false, err
		}
		for _, v := range buff {
			if v != 0 {
				data_buff = append(data_buff, v)
			}
		}
		if length < buff_size {
			break
		}
	}
	data = string(data_buff)
	// 复用server里面processCmdStrToSlice的代码
	may_cmd := bufio.NewScanner(strings.NewReader(data))
	may_cmd.Split(bufio.ScanWords)
	may_cmd_slice = make([]string, 0)
	for may_cmd.Scan() {
		may_cmd_slice = append(may_cmd_slice, may_cmd.Text())
	}
	if may_cmd_slice[0] == "[#clientmov#]" {
		command = true
	} else {
		command = false
	}

	return data, may_cmd_slice[1:], command, nil
}

func processCmd(cmd []string) (err error) { // 服务端发过来的源头上就不会有问题，因此不做检查
	switch cmd[0] {
	case "login_success":
		login_flag = true
		current_id = cmd[2]
	default:
		fmt.Println("Server command not found")
	}
	return nil
}
