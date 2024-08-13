package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type MtrHop struct {
	IP       []string
	Host     []string
	ASN      []string
	Timings  []time.Duration //In Json they become nanosecond
	Avg      time.Duration
	Loss     int
	SD       time.Duration
	Sent     int
	Received int
	Last     time.Duration
	Best     time.Duration
	Worst    time.Duration
}

// Summarize calculates various statistics for each Hop
func (hop *MtrHop) Summarize(count int) {
	//After the Timings block has been populated.
	hop.Sent = count
	hop.Received = len(hop.Timings)
	if len(hop.Timings) > 0 {
		hop.Last = hop.Timings[len(hop.Timings)-1]
		hop.Best = hop.Timings[0]
		hop.Worst = hop.Timings[0]
	}
	for _, t := range hop.Timings {
		hop.Avg += t / time.Duration(hop.Received)
		if hop.Best > t {
			hop.Best = t
		}
		if hop.Worst < t {
			hop.Worst = t
		}
	}
	hop.SD = stdDev(hop.Timings, hop.Avg)
	hop.Loss = (100 * (hop.Sent - hop.Received)) / hop.Sent
	//Populate ips
	hop.ResolveIPs()
}

// ResolveIPs populates the DNS hostnames of the IP in each Hop.
func (hop *MtrHop) ResolveIPs() {
	hop.Host = make([]string, len(hop.IP))
	for idx, ip := range hop.IP {
		hop.Host[idx] = reverselookup(ip)
	}
}

type rawhop struct {
	datatype string
	idx      int
	value    string
	value2   string
}

type MTROutPut struct {
	Hops     []*MtrHop
	Target   string //Saving this FYI
	HopCount int
}

// Summarize calls Summarize on each Hop
func (result *MTROutPut) Summarize(count int) {
	for _, hop := range result.Hops {
		hop.Summarize(count)
	}
}

// Calculate per hop RTT difference to the previous hop
func (result *MTROutPut) Diff(count int) {
	hopCount := result.HopCount - 1

	for i := 0; i < hopCount; i++ {
		if len(result.Hops[i].IP) == 0 {
			fmt.Println("Hop: *")
		} else {
			fmt.Println(result.Hops[i].IP[0])
			if len(result.Hops[i].Timings) != count {
				fmt.Printf("expected %d timings, got %d for hop: %s\n", count, len(result.Hops[i].Timings), result.Hops[i].IP[0])
			}
			fmt.Print(result.Hops[i].Timings)
		}
		fmt.Println()
	}

	// create a 2D slice with hopCount rows and count columns to store the difference
	diff := make([][]time.Duration, hopCount)
	for i := 0; i < hopCount; i++ {
		diff[i] = make([]time.Duration, count)
		for j := 0; j < count; j++ {
			diff[i][j] = 0
		}
	}

	// calculate the difference from the current hop to the previous hop
	for i := 0; i < hopCount; i++ {
		for j := 0; j < count; j++ {
			if i > 0 {
				if len(result.Hops[i-1].Timings) > 0 {
					diff[i][j] = result.Hops[i].Timings[j] - result.Hops[i-1].Timings[j]
				} else {
					diff[i][j] = result.Hops[i].Timings[j]
				}
			}
		}
	}

	// print the difference
	fmt.Println("\nRTT difference to the previous hop:")
	for i := 0; i < hopCount; i++ {
		fmt.Print(diff[i])
		fmt.Println()
	}
}

// String returns output similar to --report option in mtr
func (result *MTROutPut) String() string {
	output := fmt.Sprintf("HOST: %sLoss%%   Snt   Last   Avg  Best  Wrst StDev", trimpad("TODO hostname", 40))
	for i, hop := range result.Hops {
		h := "???"
		if len(hop.IP) > 0 {
			//fmt.Println(hop.Host, hop.IP)
			h = hop.Host[0]
			if h == "" {
				h = hop.IP[0]
			}
		}
		output = fmt.Sprintf("%s\n%2d.|-- %s%3d.0%%   %3d  %5.1f %5.1f %5.1f %5.1f %5.1f", output, i, trimpad(h, 38), hop.Loss, hop.Sent, durms(hop.Last), durms(hop.Avg), durms(hop.Best), durms(hop.Worst), durms(hop.SD))
	}
	return output
}

// NewMTROutPut can be used to parse output of mtr --raw <target ip> .
// raw is the output from mtr command, count is the -c argument, default 10 in mtr
func NewMTROutPut(raw, target string, count int) (*MTROutPut, error) {
	//last hop comes in multiple times... https://github.com/traviscross/mtr/blob/master/FORMATS
	out := &MTROutPut{}
	out.Target = target
	rawhops := make([]rawhop, 0)
	//Store each line of output in rawhop structure
	for _, line := range strings.Split(raw, "\n") {
		things := strings.Split(line, " ")
		if len(things) == 3 || (len(things) == 4 && things[0] == "p") {
			idx, err := strconv.Atoi(things[1])
			fmt.Println(things)
			if err != nil {
				return nil, err
			}
			data := rawhop{
				datatype: things[0],
				idx:      idx,
				value:    things[2],
			}
			if len(things) == 4 {
				data.value2 = things[3]
			}
			rawhops = append(rawhops, data)
			//Number of hops = highest index+1
			if out.HopCount < (idx + 1) {
				out.HopCount = idx + 1
			}
		}
	}
	out.Hops = make([]*MtrHop, out.HopCount)
	for idx, _ := range out.Hops {
		out.Hops[idx] = &MtrHop{
			Timings: make([]time.Duration, 0),
			IP:      make([]string, 0),
			Host:    make([]string, 0),
		}
	}

	/*
		hostline:
		h <pos> <host IP>

		xmitline:
		x <pos> <seqnum>

		pingline:
		p <pos> <pingtime (ms)> <seqnum>

		dnsline:
		d <pos> <hostname>

		timestampline:
		t <pos> <pingtime> <timestamp>

		mplsline:
		m <pos> <label> <traffic_class> <bottom_stack> <ttl>
	*/

	/*
		stats = {
			"hop1": {
				"30001": 0,
				"30002": 0,
			},
			"hop2": {
				"30003": 0,
				"30004": 0,
			},
		}
	*/

	stats := make(map[string]map[string]int)

	for _, data := range rawhops {
		switch data.datatype {
		case "h":
			out.Hops[data.idx].IP = append(out.Hops[data.idx].IP, data.value)
		case "x":
			// xmitline
			stats[data.value][data.value2] = 0
		//case "d":
		//Not entirely sure if multiple IPs. Better use -n in mtr and resolve later in summarize.
		//out.Hops[data.idx].Host = append(out.Hops[data.idx].Host, data.value)
		case "p":
			t, err := strconv.Atoi(data.value)
			if err != nil {
				return nil, err
			}
			stats[data.value][data.value2] = t
			out.Hops[data.idx].Timings = append(out.Hops[data.idx].Timings, time.Duration(t)*time.Microsecond)
		}
	}
	//Filter dupe last hops
	finalidx := 0
	previousip := ""
	for idx, hop := range out.Hops {
		if len(hop.IP) > 0 {
			if hop.IP[0] != previousip {
				previousip = hop.IP[0]
				finalidx = idx + 1
			}
		}
	}
	out.Hops = out.Hops[0:finalidx]
	return out, nil
}

// Execute mtr command and return parsed output
func ExecuteMTR(target string, IPv string, count int, interval float64) (*MTROutPut, error) {
	return ExecuteMTRContext(context.Background(), target, IPv, count, interval)
}

// Execute mtr command and return parsed output,
// killing the process if context becomes done before command completes.
func ExecuteMTRContext(ctx context.Context, target string, IPv string, count int, interval float64) (*MTROutPut, error) {
	//Validate r.Target before sending
	tgt := strings.Trim(target, "\n \r") //Trim whitespace
	if strings.Contains(tgt, " ") {      //Ensure it doesnt contain space
		return nil, errors.New("invalid hostname")
	}
	if strings.HasPrefix(tgt, "-") { //Ensure it doesnt start with -
		return nil, errors.New("invalid hostname")
	}
	realtgt := ""
	parsed := net.ParseIP(tgt)
	if parsed != nil {
		realtgt = tgt
	}
	var addrs []net.IP
	var err error
	if realtgt == "" {
		addrs, err = net.LookupIP(tgt)
		if err != nil {
			return nil, err
		}
		if len(addrs) == 0 {
			return nil, errors.New("host not found")
		}
	}
	var cmd *exec.Cmd

	switch IPv {
	case "4":
		if realtgt == "" {
			for _, ip := range addrs {
				i := ip.To4()
				if i != nil {
					realtgt = i.String()
				}
			}
		}
		if realtgt == "" {
			return nil, errors.New("no IPv4 address found")
		}
		cmd = exec.CommandContext(ctx, "mtr", "--raw", "-n", "-c", fmt.Sprintf("%d", count), "-4", "-i", fmt.Sprintf("%f", interval), realtgt)
	case "6":
		if realtgt == "" {
			for _, ip := range addrs {
				i := ip.To16()
				if i != nil && ip.To4() == nil { //Explicitly check if its not v4
					realtgt = i.String()
				}
			}
		}
		if realtgt == "" {
			return nil, errors.New("no IPv6 address found")
		}
		cmd = exec.CommandContext(ctx, "mtr", "--raw", "-n", "-c", fmt.Sprintf("%d", count), "-6", "-i", fmt.Sprintf("%f", interval), realtgt)
	default:
		if realtgt == "" {
			if doesipv6() {
				for _, ip := range addrs {
					i := ip.To16()
					if i != nil {
						realtgt = i.String()
					}
				}
			}
			if realtgt == "" {
				for _, ip := range addrs {
					i := ip.To4()
					if i != nil {
						realtgt = i.String()
					}
				}
			}
		}
		if realtgt == "" {
			return nil, errors.New("no IP address found")
		}
		//realtgt = addrs[0].String() //Choose first addr..
		cmd = exec.CommandContext(ctx, "mtr", "--raw", "-n", "-c", fmt.Sprintf("%d", count), "-i", fmt.Sprintf("%f", interval), realtgt)
	}

	fmt.Println(cmd.String())
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		if ctx.Err() != nil {
			// Process was killed by context
			return nil, ctx.Err()
		}
		// Process finished with error code
		return nil, errors.New(stderr.String())
	}
	return NewMTROutPut(out.String(), realtgt, 10)
}
