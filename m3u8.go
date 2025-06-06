package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
)

var Lock = sync.Map{}

func GetM3U8(instance *WrapperInstance, adamId string) (string, error) {
	lock, _ := Lock.LoadOrStore(instance.Id, &sync.Mutex{})
	lock.(*sync.Mutex).Lock()
	defer lock.(*sync.Mutex).Unlock()
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", instance.M3U8Port))
	if err != nil {
		return "", err
	}
	defer conn.Close()
	_, err = conn.Write([]byte{byte(len(adamId))})
	if err != nil {
		return "", err
	}
	_, err = io.WriteString(conn, adamId)
	if err != nil {
		return "", err
	}
	response, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		return "", err
	}
	if len(response) > 0 {
		response = bytes.TrimSpace(response)
		return string(response), nil

	} else {
		return "", errors.New("empty response")
	}
}
