package main

import (
	"encoding/binary"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"sync"
	"sync/atomic"
)

var DefaultId = "0"
var PrefetchKey = "skd://itunes.apple.com/P000000000/s1/e1"

type DecryptInstance struct {
	Id          string
	Region      string
	DecryptPort int
	M3U8Port    int
	LastAdamId  string
	LastKey     string
	busy        int32 // 使用 atomic 操作
	conn        net.Conn
	mu          sync.Mutex // 保护状态字段
}

func (d *DecryptInstance) Handle(task *Task) {
	if !atomic.CompareAndSwapInt32(&d.busy, 0, 1) {
		log.Warnf("Handle called on busy instance %s\n", d.Id)
		return
	}

	defer func() {
		atomic.StoreInt32(&d.busy, 0)
		go DispatcherInstance.OnInstanceFree(d)
	}()

	d.mu.Lock()
	var refresh int
	if task.Key == d.LastKey {
		refresh = 0
	} else if d.LastAdamId == "" && d.LastKey == "" {
		refresh = 1
	} else {
		refresh = 2
	}

	d.LastAdamId = task.AdamId
	d.LastKey = task.Key
	d.mu.Unlock()

	result, err := decrypt(d.conn, task.AdamId, task.Payload, task.Key, refresh)

	select {
	case task.Result <- &Result{Data: result, Error: err}:
	default:
		log.Warnf("Failed to send result for task %s:%s\n", task.AdamId, task.Key)
	}
}

func (d *DecryptInstance) SetBusy(b bool) {
	if b {
		atomic.StoreInt32(&d.busy, 1)
	} else {
		atomic.StoreInt32(&d.busy, 0)
	}
}

func (d *DecryptInstance) IsBusy() bool {
	return atomic.LoadInt32(&d.busy) != 0
}

func decrypt(conn net.Conn, adamId string, sample []byte, key string, refresh int) ([]byte, error) {
	if refresh != 0 {
		if key == PrefetchKey {
			adamId = DefaultId
		}
		if refresh == 2 {
			_, err := conn.Write([]byte{0, 0, 0, 0})
			if err != nil {
				return nil, err
			}
		}
		_, err := conn.Write([]byte{byte(len(adamId))})
		if err != nil {
			return nil, err
		}
		_, err = io.WriteString(conn, adamId)
		if err != nil {
			return nil, err
		}
		_, err = conn.Write([]byte{byte(len(key))})
		if err != nil {
			return nil, err
		}
		_, err = io.WriteString(conn, key)
		if err != nil {
			return nil, err
		}
	}
	err := binary.Write(conn, binary.LittleEndian, uint32(len(sample)))
	if err != nil {
		return nil, err
	}
	_, err = conn.Write(sample)
	if err != nil {
		return nil, err
	}
	de := make([]byte, len(sample))
	_, err = io.ReadFull(conn, de)
	if err != nil {
		return nil, err
	}
	return de, nil
}
