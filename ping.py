import os
import time
import json
from math import fabs
from pprint import pprint
from pathlib import Path
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


def boxplot_position():
    positions = {}
    for i in range(1, len(hops)):
        key = "{}\n-> {}".format(hops[i-1], hops[i])
        positions[key] = len(hops) - i
    return positions

result = []


def do_ping():
    end_time = datetime.now() + timedelta(seconds=14)

    with open("ping-{}.json".format(datetime.now().strftime("%Y%m%d-%H%M%S")), "w") as f:
        while True:
            time.sleep(0.01)
            if datetime.now() >= end_time:
                break

            current_round = dict.fromkeys(hops, 0)
            for hop in hops:
                #print("Now: {}, Pinging {}...".format(datetime.now(), hop))
                host = ping(hop, count=1, timeout=0.5)
                if host.min_rtt == 0:
                    break
                current_round[hop] = host.min_rtt
                #print(host.min_rtt)
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

    roundCount = 0
    path = Path("./latency")
    for dirpath, dirnames, files in os.walk(path):
        if len(files) != 0:
            for f in files:
                if f.endswith(".json") and f.startswith("ping"):
                    roundCount += 1
                    # if roundCount > 1:
                        # break
                    filename = os.path.join(dirpath, f)
                    print(filename)
                    with open(filename, "r") as jsonfile:
                        data = json.load(jsonfile)
                        for round in data:
                            for i in range(1, len(hops)):
                                key = "{}\n-> {}".format(hops[i-1], hops[i])
                                if key not in diff:
                                    diff[key] = []
                                rtt_diff = round[hops[i]] - round[hops[i-1]]
                                if rtt_diff < 0:
                                    continue
                                diff[key].append(fabs(rtt_diff))
    plt.figure(figsize=(10, 5))
    pos = boxplot_position()
    for position, column in enumerate(pos):
        print(min(diff[column]), max(diff[column]))
        plt.boxplot(diff[column], positions=[len(pos) - position], vert=False, meanline=True, meanprops={"color": "red"}, showmeans=False, showfliers=False)

    # plt.xlim(-5, 50)
    plt.xlabel("Latency (ms)")
    plt.yticks(range(1, len(diff.keys())+1), reversed(diff.keys()), rotation=0)
    plt.tight_layout()
    plt.savefig("ping.png")
    plt.close()


if __name__ == "__main__":
    # start_thread_at(do_ping, get_start_time())
    plot()
