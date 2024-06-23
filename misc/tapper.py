import requests
import random
import copy
import concurrent.futures
import re
import time
import argparse


def try_ignore(func):
    def wrapper(*args, **kwargs):
        try:
            return func(*args, **kwargs)
        except Exception as e:
            print(e)
            return None

    return wrapper


BOT_START_REQUEST = {
    "update_id": -1,
    "message": {
        "date": 1700000000,
        "chat": {
            "last_name": "Last",
            "id": -1,
            "first_name": "First",
            "username": "username",
        },
        "message_id": -1,
        "from": {
            "last_name": "Last",
            "id": -1,
            "first_name": "First",
            "username": "username",
        },
        "text": "/start",
    },
}


def get_start_request(uid=None) -> dict:
    req = copy.deepcopy(BOT_START_REQUEST)
    uid = uid or random.randint(10**6, 10**7)
    req["message"]["from"]["id"] = uid
    req["message"]["from"]["username"] = f"username_{uid}"
    req["message"]["chat"]["id"] = uid
    return req


API_TAP_REQUEST = {
    "init": "",
    "e": 64,
    "s": 6759,
    "uid": 6163587238,
    "xyz": [0] * 10,
}

API_CALLS = 0


def get_tap_request(uid=None, energy_range=(-333, 999)) -> dict:
    req = copy.deepcopy(API_TAP_REQUEST)
    req["init"] = (
        """init-validation-disabled=true&hash=673fa9a2c4b3d80d5c163ad23eb55be7"""
    )
    nrg = random.randint(*energy_range)
    if nrg == 0:
        nrg = 1
    req["uid"] = uid or random.randint(10**6, 10**7)
    req["e"] = nrg
    req["s"] = 0
    return req


def tap(uid=None, c=100) -> dict:
    global API_CALLS
    API_CALLS += 1
    req = get_tap_request(uid)
    rsp = requests.post(f"{API_BASE}/taps/", json=req)
    rsp.raise_for_status()
    return req, rsp.json()


def tapper(uid=None) -> dict:
    global API_CALLS
    API_CALLS += 1
    req = get_start_request(uid)
    rsp = requests.post(f"{API_BASE}/tghook/", json=req)
    rsp.raise_for_status()
    return req, rsp.json()


START_MSG_RE = re.compile(r"you have ([\-0-9]+) points after ([0-9]+) rounds")


def run_single_tapper(rounds, uid=None):
    ts_start = time.time()
    total_energy = 0
    uid = uid or random.randint(10**6, 10**7)
    tapper(uid)  # init tapper
    for _ in range(rounds):
        req, _ = tap(uid)
        total_energy += req["e"]
    req, rsp = tapper(uid)  # get tapper data
    ts_stop = time.time()
    mo = START_MSG_RE.search(rsp["text"])
    if mo:
        qscore, qrounds = map(int, (mo.group(1), mo.group(2)))
        rc = "+" if (total_energy == qscore and rounds == qrounds) else "-"
        print(
            f"{rc} {uid=} {total_energy}={qscore} {rounds}={qrounds} {int(ts_stop*1000-ts_start*1000)}μs"
        )
    else:
        print("!", rsp["text"])


def run_multi_tappers(runs=5, c=None, m=5):
    uid = random.randint(10**6, 10**7)
    with concurrent.futures.ThreadPoolExecutor(m) as pool:
        for i in range(runs):
            future = pool.submit(
                run_single_tapper,
                c if c is not None else random.randint(1, 50),
                uid + i,
            )
            future.add_done_callback(lambda x: True)


def read_params():
    parser = argparse.ArgumentParser()
    parser.add_argument("--api-url", type=str, default="http://127.0.0.1:8765")
    parser.add_argument("--runs", type=int, default=3)
    parser.add_argument("--taps", type=int, default=50)
    parser.add_argument("--plls", type=int, default=5)
    params = parser.parse_args()
    global API_BASE
    API_BASE = params.api_url
    return params


def main():
    params = read_params()
    print(params)

    ts_start = time.time()

    print("# single_tapper")
    for _ in range(params.runs):
        run_single_tapper(random.randint(1, params.taps))

    print("# multi_tappers")
    run_multi_tappers(params.runs, params.taps, params.plls)

    ts_stop = time.time()
    run_time = ts_stop - ts_start
    print(f"{API_CALLS=} {run_time=:.3f}s {API_CALLS/run_time=:.3f}calls/s")


if __name__ == "__main__":
    main()
