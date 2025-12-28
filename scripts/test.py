import requests
import random
import time

API_BASE_URL = "http://localhost:8081/api/v1/leaderboard"

# Simulate score submission
def submit_score(user_id):
    score = random.randint(100, 10000)
    requests.post(
        f"{API_BASE_URL}/submit",
        json={"user_id": user_id, "score": score}
    )

# Fetch top players
def get_top_players():
    requests.get(f"{API_BASE_URL}/top")

while True:
    user_id = random.randint(1, 1000000)
    submit_score(user_id)

    if random.random() < 0.3:
        get_top_players()

    time.sleep(0.05)
