package main

import (
	"fmt"
	"math"
	"net"
	"time"
)

func doesipv6() bool {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return false
	}
	localipv6 := []string{"fc00::/7", "::1/128", "fe80::/10"}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			ip6 := ipnet.IP.To16()
			ip4 := ipnet.IP.To4()
			if ip6 != nil && ip4 == nil {
				local := false
				for _, r := range localipv6 {
					_, cidr, _ := net.ParseCIDR(r)
					if cidr.Contains(ip6) {
						local = true
					}
				}
				if !local {
					return true
				}
			}
		}
	}
	return false
}

func stdDev(timings []time.Duration, avg time.Duration) time.Duration {
	//taken from https://github.com/ae6rt/golang-examples/blob/master/goeg/src/statistics_ans/statistics.go
	if len(timings) < 2 {
		return time.Duration(0)
	}
	mean := float64(avg)
	total := 0.0
	for _, t := range timings {
		number := float64(t)
		total += math.Pow(number-mean, 2)
	}
	variance := total / float64(len(timings)-1)
	std := math.Sqrt(variance)
	return time.Duration(std)
}

type lookupresult struct {
	addr []string
	err  error
}

func reverselookup(ip string) string {
	result := ""
	ch := make(chan lookupresult, 1)
	go func() {
		addr, err := net.LookupAddr(ip)
		ch <- lookupresult{addr, err}
	}()
	select {
	case res := <-ch:
		if res.err == nil {
			//log.Println(addr[0], err)
			if len(res.addr) > 0 {
				result = res.addr[0]
			}
		}
	case <-time.After(time.Second * 1):
		result = ""
	}
	return result
}

// Helper function to trim or pad a string
func trimpad(input string, size int) string {
	if len(input) > size {
		input = input[0:size]
	}
	return fmt.Sprintf("%[1]*[2]s ", size*-1, input)
}

// Return milliseconds as floating point from duration
func durms(d time.Duration) float64 {
	return float64(d.Nanoseconds()) / (1000 * 1000)
}
