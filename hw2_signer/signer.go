package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

var mutex = &sync.Mutex{}

func Md5Internal(data string) string {
	mutex.Lock()
	result := DataSignerMd5(data)
	mutex.Unlock()
	return result
}

func Crc32Internal(data string, out chan string) {
	out <- DataSignerCrc32(data)
}

func ToStringCustom(input interface{}) string {
	switch input.(type) {
	case int:
		return strconv.Itoa(input.(int))
	case string:
		return input.(string)
	default:
		panic("cant convert data to string")
	}
}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}

	for dataRaw := range in {
		dataStr := ToStringCustom(dataRaw)
		wg.Add(1)

		go func(data string, wgInt *sync.WaitGroup) {
			defer wgInt.Done()
			left := make(chan string)
			right := make(chan string)

			go Crc32Internal(data, left)
			go Crc32Internal(Md5Internal(data), right)

			result := ""
			fromLeft := false
			fromRight := false
		LOOP:
			for {
				select {
				case l := <-left:
					fromLeft = true
					result = l + "~" + result
				case r := <-right:
					fromRight = true
					result = result + r
				default:
					if fromLeft && fromRight {
						break LOOP
					}
				}
			}
			out <- result
		}(dataStr, wg)
	}

	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for dataRaw := range in {
		dataStr := ToStringCustom(dataRaw)
		wg.Add(1)

		go func(data string, wgInt *sync.WaitGroup) {
			defer wgInt.Done()

			var chans [6]chan string
			for i := range chans {
				chans[i] = make(chan string, 1)
				go Crc32Internal(strconv.Itoa(i)+data, chans[i])
			}
			result := ""
			for i := range chans {
				val := <-chans[i]
				result = result + val
			}

			out <- result
		}(dataStr, wg)
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	values := []string{}
	for dataRaw := range in {
		dataStr := ToStringCustom(dataRaw)
		values = append(values, dataStr)
	}
	sort.Strings(values)
	result := strings.Join(values, "_")
	out <- result
}

func ExecutePipeline(jobs ...job) {
	var previous chan interface{}
	for i, j := range jobs {
		out := make(chan interface{}, 1)

		if i == len(jobs)-1 {
			j(previous, out)
			close(out)
			return
		}

		go func(jobInt job, i, o chan interface{}) {
			jobInt(i, o)
			close(o)
		}(j, previous, out)

		previous = out
	}
}
