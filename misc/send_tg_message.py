import os
import sys
import argparse
import requests


def send_tg_message(token, chat_id, message):
    url = f"https://api.telegram.org/bot{token}/sendMessage"
    payload = {"chat_id": chat_id, "text": message}
    response = requests.post(url, json=payload)
    try:
        data = response.json()
    except:
        data = response.text
    return response.status_code, data


def read_args():
    parser = argparse.ArgumentParser(description="send tg message")
    parser.add_argument("--bot-token", default=os.getenv("QQOIN_BOT_TOKEN"))
    parser.add_argument("--chat-ids", required=True)
    parser.add_argument("--message", required=True)
    args = parser.parse_args()

    assert args.bot_token, "no bot token"
    assert args.chat_ids, "no chat ids"
    assert all(c.isdigit() for c in args.chat_ids.split(",")), "bad chat-ids format"
    assert args.message, "no message"

    return args


def main():
    args = read_args()
    for chat_id in args.chat_ids.split(","):
        response = send_tg_message(args.bot_token, int(chat_id), args.message)
        print(f"{chat_id=} {response=}")


if __name__ == "__main__":
    try:
        main()
    except Exception as e:
        print("error:", e, file=sys.stderr)
        sys.exit(1)
