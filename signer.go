package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// сюда писать код

func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{})

	for _, j := range jobs {
		wg.Add(1)
		out := make(chan interface{})

		go func(job job, in, out chan interface{}, wg *sync.WaitGroup) {
			defer wg.Done()
			defer close(out)
			job(in, out)
		}(j, in, out, wg)
		in = out
	}
	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}

	for i := range in {
		time.Sleep(time.Millisecond * 11)
		data := i.(int)
		wg.Add(1)

		go func(data int) {
			strData := strconv.Itoa(data)
			var strCrc32, strMd5 string
			wgLocal := &sync.WaitGroup{}
			wgLocal.Add(2)

			go func(strData string) {
				strCrc32 = DataSignerCrc32(strData)
				wgLocal.Done()
			}(strData)

			go func(strData string) {
				strMd5 = DataSignerCrc32(DataSignerMd5(strData))
				wgLocal.Done()
			}(strData)

			wgLocal.Wait()
			out <- strCrc32 + "~" + strMd5
			wg.Done()
		}(data)
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	const th int = 6
	wg := &sync.WaitGroup{}

	for i := range in {
		wg.Add(1)

		go func(in string, out chan interface{}, th int, wg *sync.WaitGroup) {
			defer wg.Done()
			mu := &sync.Mutex{}
			wgLocal := &sync.WaitGroup{}
			result := make([]string, th)

			for i := 0; i < th; i++ {
				wgLocal.Add(1)
				data := strconv.Itoa(i) + in

				go func(acc []string, index int, data string, jobWg *sync.WaitGroup, mu *sync.Mutex) {
					defer wgLocal.Done()
					data = DataSignerCrc32(data)

					mu.Lock()
					acc[index] = data
					mu.Unlock()
				}(result, i, data, wgLocal, mu)
			}
			wgLocal.Wait()
			out <- strings.Join(result, "")
		}(i.(string), out, th, wg)
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var result []string

	for i := range in {
		result = append(result, i.(string))
	}
	sort.Strings(result)
	out <- strings.Join(result, "_")
}
