import time
import json
from pprint import pprint
from datetime import datetime, timedelta
from multiprocessing import Process
from threading import Thread

from icmplib import ping
from matplotlib import pyplot as plt


hops = [
    "192.168.1.1",
    "100.64.0.1",
    "149.19.109.47",
    "172.217.18.100"
]


result = []

def do_ping():
    end_time = datetime.now() + timedelta(seconds=14)

    with open("ping-{}.json".format(datetime.now().strftime("%Y%m%d-%H%M%S")), "w") as f:
        while True:
            if datetime.now() >= end_time:
                break

            current_round = dict.fromkeys(hops, 0)
            for hop in hops:
                print("Now: {}, Pinging {}...".format(datetime.now(), hop))
                host = ping(hop, count=1, timeout=0.5)
                if host.min_rtt == 0:
                    break
                current_round[hop] = host.min_rtt
                print(host.min_rtt)
                time.sleep(0.01)
            result.append(current_round)

        json.dump(result, f, indent=2)


def get_start_time():
    now = datetime.now()
    if now.second >= 12 and now.second < 27:
        return now.replace(second=27).replace(microsecond=0)
    elif now.second >= 27 and now.second < 42:
        return now.replace(second=42).replace(microsecond=0)
    elif now.second >= 42 and now.second < 57:
        return now.replace(second=57).replace(microsecond=0)
    elif now.second >= 0 and now.second < 12:
        return now.replace(second=12).replace(microsecond=0)
    else:
        return now.replace(second=12).replace(microsecond=0) + timedelta(minutes=1)


def start_thread_at(func, target_time):
    delay = (target_time - datetime.now()).total_seconds()
    if delay > 0:
        print("Waiting for {} seconds...".format(delay))
        time.sleep(delay)

    th = Thread(target=func)
    th.start()
    th.join()


def plot():
    diff = {}
    with open("ping.json", "r") as f:
        data = json.load(f)
        for round in data:
            print(round)
            # calculate the difference between each hop to their previous hop
            for i in range(1, len(hops)):
                key = "{} -> {}".format(hops[i-1], hops[i])
                if key not in diff:
                    diff[key] = []
                rtt_diff = round[hops[i]] - round[hops[i-1]]
                if rtt_diff < 0:
                    rtt_diff = 0
                diff[key].append(rtt_diff)
    pprint(diff)
    # boxplot the difference
    fig = plt.figure(figsize=(15, 10))
    plt.boxplot(diff.values())
    plt.xticks(range(1, len(diff.keys())+1), diff.keys(), rotation=0)
    plt.tight_layout()
    plt.savefig("ping.png")
    plt.close()


if __name__ == "__main__":
    start_thread_at(do_ping, get_start_time())
    plot()
