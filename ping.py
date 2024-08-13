import time
from datetime import datetime, timedelta
from multiprocessing import Process

from icmplib import ping


hops = [
    "192.168.1.1",
    "100.64.0.1",
    "172.16.248.4",
    "149.19.109.14",
    "210.171.224.96",
    "142.251.233.111",
    "142.250.224.213",
    "142.250.196.142"
]

# hops = [
#     "192.168.1.254",
#     "10.31.36.1",
#     "154.11.2.254",
#     "154.11.15.73",
#     "1.1.1.1"
# ]

def do_ping():
    while True:
        for hop in hops:
            print("Now: {}, Pinging {}...".format(datetime.now(), hop))
            host = ping(hop, count=1)
            print(host.min_rtt)
            time.sleep(0.01)


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

    process = Process(target=func)
    process.start()
    process.join(14)
    process.terminate()


if __name__ == "__main__":
    start_thread_at(do_ping, get_start_time())
