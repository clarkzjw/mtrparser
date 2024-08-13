# hostline:
# h <pos> <host IP>

# xmitline:
# x <pos> <seqnum>

# pingline:
# p <pos> <pingtime (ms)> <seqnum>

from pprint import pprint

with open("res", "r") as f:
    lines = f.readlines()

    result = {}

    for line in lines:
        print(line.strip())

        current = line.strip().split(" ")
        package_type = current[0]
        hop = current[1]
        if hop == "0":
            continue

        if hop not in result.keys():
            result[hop] = {
                "ip": [],
                "rtt": {}
            }

        if package_type == "h":
            ip = current[2]

            result[hop]["ip"].append(ip)
        if package_type == "x":
            seqnum = current[2]
            result[hop]["rtt"][seqnum] = None
            # result[hop]["rtt"][]{
            #     "seq": seqnum,
            #     "pingtime": None
            # }
        if package_type == "p":
            pingtime = current[2]
            seqnum = current[3]
            result[hop]["rtt"][seqnum] = pingtime

        # result.append({
        #     hop: {
        #         "ip": ip,
        #         "seqnum": seqnum,
        #         "pingtime": pingtime
        #     }
        # })
    pprint(result)
