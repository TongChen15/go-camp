package main

import (
	"bytes"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/tjarratt/babble"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

var (
	pool  = newPool()
	sizes = []int32{10, 20, 50, 100, 200, 500, 1000, 2000, 5000}
)

func main() {
	client := pool.Get()
	defer client.Close()

	// Problem #2
	keyer := babble.NewBabbler()
	valuer := babble.NewBabbler()
	keyer.Separator = "_"
	valuer.Separator = " "
	keyer.Count = 3   // 3 random words to make a key
	valuer.Count = 10 // 10 random words to make a value

	err := ioutil.WriteFile("redis-memory-blueprint.txt", []byte("file structure contains blocks with Key size and memory information\n"), 0644)
	if err != nil {
		log.Fatal(err)
	}
	file2, err := os.OpenFile("redis-memory-blueprint.txt", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file2.Close()

	for i := 0; i <= 500000; i++ {
		if i%10000 == 0 {
			if _, err := file2.WriteString("*********************************\n\n"); err != nil {
				log.Fatal(err)
			}
			keyNum, _ := client.Do("info", "keyspace")
			str := keyNum.([]uint8)
			if _, err := file2.WriteString(string(str)); err != nil {
				log.Fatal(err)
			}
			memInfo, _ := client.Do("info", "memory")
			str = memInfo.([]uint8)
			if _, err := file2.WriteString(string(str)); err != nil {
				log.Fatal(err)
			}

		}
		key := keyer.Babble()
		value := valuer.Babble()
		// fmt.Printf("\nkey: %s \nvalue: %s\n", key, value)
		_, err := client.Do("SET", key, value)
		if err != nil {
			panic(err)
		}
	}

	// Problem #1
	err = ioutil.WriteFile("redis-set-get-perf.txt", []byte("file contains redis get set performance information\n"), 0644)
	if err != nil {
		log.Fatal(err)
	}
	file1, err := os.OpenFile("redis-set-get-perf.txt", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file1.Close()

	for _, size := range sizes {
		size := fmt.Sprintf("%d", size)
		cmd := exec.Command("redis-benchmark", "-t", "set,get", "-c", "1000", "-P", "10", "-d", size)

		heading := fmt.Sprintf("***** command %v is running *****\n", cmd.Args)
		fmt.Print(heading)
		if _, err := file1.WriteString(heading); err != nil {
			log.Fatal(err)
		}

		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
			return
		}
		if _, err := file1.WriteString(out.String()); err != nil {
			log.Fatal(err)
		}
	}
}

func newPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:   80,
		MaxActive: 12000,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "127.0.0.1:6379")
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
}
