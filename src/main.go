package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

type USER struct {
	conn       *net.TCPConn
	login_flag bool
	current_id string
	chat_mode  bool
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
func getInput() (input_str string) {
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	input_str = input.Text()
	return input_str
}
func keepRecvDataAndProcess(user *USER) {
	for true {
		buff_size := 128
		var data_buff []byte
		for true { // 缓冲区
			buff := make([]byte, buff_size)
			length, err := user.conn.Read(buff)
			if err != nil {
				panic(err)
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
		data_str := string(data_buff)
		// 字符串收到
		cmd_slice := make([]string, 0)
		command := bufio.NewScanner(strings.NewReader(data_str))
		command.Split(bufio.ScanWords)
		for command.Scan() {
			cmd_slice = append(cmd_slice, command.Text())
		}
		if cmd_slice[0] == "[#clientmov#]" { // 是来自server的command
			cmd_slice = cmd_slice[1:]
			processCmd(cmd_slice, user)
		} else { // 不是command，直接输出
			fmt.Print(data_str)
		}
	}
}

var user USER

func main() {
	user.conn = connectServer("127.0.0.1:5000")
	defer user.conn.Close()
	go keepRecvDataAndProcess(&user)
	sendCommand("sendprompt", &user) // 获取第一个提示符
	for true {
		input_data := getInput()
		sendCommand(input_data, &user)
	}

}

func sendCommand(data_str string, user *USER) {
	// 部分command需要登录后才可以使用
	cmd_slice := make([]string, 0)
	tmp := bufio.NewScanner(strings.NewReader(data_str))
	tmp.Split(bufio.ScanWords)
	for tmp.Scan() {
		cmd_slice = append(cmd_slice, tmp.Text())
	}
	if len(cmd_slice) != 0 { // 只有空格导致越界
		switch cmd_slice[0] {
		case "startchat":
			fallthrough
		case "sendmsg":
			fallthrough
		case "checkmsg":
			if user.login_flag == false {
				fmt.Println("Loooooogin required, wanna join?")
				user.conn.Write([]byte("sendprompt"))
			} else {
				user.conn.Write([]byte(data_str))
			}
		default: // 不需要login统统不写，进入default分支
			user.conn.Write([]byte(data_str))
		}
	} else {
		if user.chat_mode == false {
			user.conn.Write([]byte("sendprompt"))
		} else {

		}
	}

}

func processCmd(cmd []string, user *USER) { // 服务端发过来的源头上就不会有问题，因此不做检查
	switch cmd[0] {
	case "login_success":
		user.login_flag = true
		user.current_id = cmd[2]
	case "send_noresponse":
		user.conn.Write([]byte("noresponse"))
	case "switch_to_chat_mode": // 由于之前是卡在getinput里面的，因此这里修改完之后，getinput收到回车才会送到send函数里面去，这时候chat_mode已经是true了
		user.chat_mode = true
		fmt.Println(user.current_id, "switched")
	case "switch_off_chat_mode":
		user.chat_mode = false
	default:
		fmt.Println("Server command not found: ", cmd[0])
	}
}
