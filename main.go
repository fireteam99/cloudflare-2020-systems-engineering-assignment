package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)


func main() {
	urlPtr := flag.String("url", "", "a valid URL (required)")
	profilePtr := flag.Int("profile", 0, "number of requests")

	flag.Parse()

	if *profilePtr < 0 {
		fmt.Println("Error: --profile takes a positive integer")
		return
	}

	if *urlPtr == "" {
		fmt.Println("Error: no URL specified")
		return
	}

	//fmt.Println("url:", *urlPtr)
	//fmt.Println("profile:", *profilePtr)

	// validate the url
	parsedURL, err := url.ParseRequestURI(*urlPtr)
	checkError(err)

	//fmt.Println("hostname:", parsedURL.Host)
	//fmt.Println("path:", parsedURL.Path)

	tcpAddr, err := net.ResolveTCPAddr("tcp4", parsedURL.Host + ":80")
	checkError(err)

	errorCodes := make(map[string]int)
	failedCount := 0
	var times []int64
	var sizes []int64

	i := 0
	for {
		start := time.Now()
		conn, err := net.DialTCP("tcp", nil, tcpAddr)
		checkError(err)

		req := fmt.Sprintf("GET %s HTTP/1.0\r\nHost: %s\r\n\r\n", parsedURL.Path, parsedURL.Host)
		_, err = conn.Write([]byte(req))
		checkError(err)

		result, err := ioutil.ReadAll(conn)
		checkError(err)
		sizes = append(sizes, int64(len(result)))

		response := string(result)
		parsedResponse := strings.Split(response, "\n")

		if len(parsedResponse) <= 0 {
			fmt.Println("Error: no response received")
		} else {
			statusLine := parsedResponse[0]
			//fmt.Println("Status Line:", statusLine)
			parsedStatusLine := strings.Split(statusLine, " ")

			code := parsedStatusLine[1]
			runes := []rune(statusLine)
			offset := utf8.RuneCountInString(parsedStatusLine[0]) + utf8.RuneCountInString(parsedStatusLine[1]) + 1
			statusMessage := strings.TrimSpace(string(runes[offset:]))
			fmt.Println("status message:",  statusMessage)

			codeInt, err := strconv.Atoi(code)
			checkError(err)
			if codeInt > 200 || codeInt < 200 {
				failedCount++
				statusCode := code + " " + statusMessage
				if _, ok := errorCodes[statusCode]; ok {
					errorCodes[statusCode] = errorCodes[statusCode] + 1
				} else {
					errorCodes[statusCode] = 1
				}
			}

		}

		conn.Close()
		duration := time.Since(start)
		times = append(times, duration.Milliseconds())
		i++

		if *profilePtr == 0 {
			fmt.Println(response)
			return
		}

		if i >= *profilePtr {
			break
		}
	}
	//fmt.Printf("%v\n", times)
	//fmt.Printf("%v\n", sizes)

	fmt.Println("Requests made:", *profilePtr)
	printTimes(times)
	printErrors(errorCodes, *profilePtr)
	printSizes(sizes)
}

func printTimes(times []int64) {
	fmt.Println("Fastest time:", min(times), "ms")
	fmt.Println("Slowest time:", max(times), "ms")
	fmt.Println("Mean time:", mean(times), "ms")
	fmt.Println("Median time:",median(times), "ms")
}

func printSizes(sizes []int64) {
	fmt.Println("Smallest response:", min(sizes), "bytes")
	fmt.Println("Largest response:",max(sizes), "bytes")
}

func removeIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}

func max(nums []int64) int64 {
	if len(nums) < 0 {
		panic("Slice is empty")
	}
	max := int64(math.MinInt64)
	for _, e := range nums {
		if e > max {
			max = e
		}
	}
	return max
}

func min(nums []int64) int64 {
	if len(nums) < 0 {
		panic("Slice is empty")
	}
	min := int64(math.MaxInt64)
	for _, e := range nums {
		if e < min {
			min = e
		}
	}
	return min
}

func mean(nums []int64) float64 {
	if len(nums) < 0 {
		panic("Slice is empty")
	}
	total := 0.0
	for _, e := range nums {
		total += float64(e)
	}
	return total / float64(len(nums))
}

func median(nums []int64) float64 {
	if len(nums) < 0 {
		panic("Slice is empty")
	}
	sort.Slice(nums, func(i, j int) bool { return nums[i] < nums[j] })
	mid := len(nums) / 2
	if len(nums) % 2 == 0 {
		return (float64(nums[mid - 1]) + float64(nums[mid])) / 2
	}
	return float64(nums[mid])
}

func printErrors(m map[string]int, attempts int) {
	errorCount := 0
	for _, v := range m {
		errorCount += v
	}

	if len(m) > 0 {
		fmt.Println("Errors:")
		for k, v := range m {
			fmt.Printf("%s x %d\n", k, v)
		}
	}
	percentage := fmt.Sprintf("%.2f%%", (float64(attempts) - float64(errorCount)) / float64(attempts) * float64(100))
	fmt.Println("Success Rate:", percentage)
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
}

