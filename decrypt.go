package main

import (
	"fmt"
	"net"
	"sync"
)

var DispatcherInstance *Dispatcher

type Dispatcher struct {
	mu        sync.RWMutex
	buckets   map[string]map[string][]*Task
	instances []*DecryptInstance
}

type Task struct {
	AdamId  string
	Key     string
	Payload []byte
	Result  chan *Result
}

type Result struct {
	Data  []byte
	Error error
}

func (d *Dispatcher) OnInstanceFree(inst *DecryptInstance) {
	if inst.IsBusy() {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	var taskToHandle *Task
	if bucketMap, exists := d.buckets[inst.LastAdamId]; exists {
		if tasks, keyExists := bucketMap[inst.LastKey]; keyExists && len(tasks) > 0 {
			taskToHandle = tasks[0]
			d.buckets[inst.LastAdamId][inst.LastKey] = tasks[1:]
		} else {
			for k, tasks := range bucketMap {
				if k != inst.LastKey && len(tasks) > 0 {
					taskToHandle = tasks[0]
					d.buckets[inst.LastAdamId][k] = tasks[1:]
					break
				}
			}
		}
	}
	if taskToHandle == nil {
		for aid, keys := range d.buckets {
			if aid == inst.LastAdamId {
				continue
			}
			for keyName, tasks := range keys {
				if len(tasks) > 0 {
					taskToHandle = tasks[0]
					d.buckets[aid][keyName] = tasks[1:]
					goto found
				}
			}
		}
	}

found:
	if taskToHandle != nil {
		go inst.Handle(taskToHandle)
	}
}

func (d *Dispatcher) AddTask(task *Task) {
	d.mu.Lock()
	if d.buckets == nil {
		d.buckets = make(map[string]map[string][]*Task)
	}
	if _, ok := d.buckets[task.AdamId]; !ok {
		d.buckets[task.AdamId] = make(map[string][]*Task)
	}
	d.buckets[task.AdamId][task.Key] = append(d.buckets[task.AdamId][task.Key], task)

	var freeInstances []*DecryptInstance
	for _, instance := range d.instances {
		if !instance.IsBusy() {
			freeInstances = append(freeInstances, instance)
		}
	}
	d.mu.Unlock()

	for _, instance := range freeInstances {
		go d.OnInstanceFree(instance)
	}
}

func (d *Dispatcher) AddInstance(inst WrapperInstance) {
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", inst.DecryptPort))
	if err != nil {
		panic(err)
	}

	decryptInst := &DecryptInstance{
		Id:          inst.Id,
		Region:      inst.Region,
		DecryptPort: inst.DecryptPort,
		M3U8Port:    inst.M3U8Port,
		LastAdamId:  "",
		LastKey:     "",
		conn:        conn,
		busy:        0,
		mu:          sync.Mutex{},
	}

	d.mu.Lock()
	d.instances = append(d.instances, decryptInst)
	d.mu.Unlock()
}

func (d *Dispatcher) RemoveInstance(inst WrapperInstance) {
	d.mu.Lock()
	defer d.mu.Unlock()

	for i, instance := range d.instances {
		if instance.Id == inst.Id {
			d.instances = append(d.instances[:i], d.instances[i+1:]...)
			break
		}
	}
}
