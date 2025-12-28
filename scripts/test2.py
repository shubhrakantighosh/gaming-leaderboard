import requests
import random
import time
import json
from datetime import datetime
from typing import List, Dict

API_BASE_URL = "http://localhost:8081/api/v1/leaderboard"

class LeaderboardTester:
    def __init__(self, base_url: str, worker_interval_minutes: int = 3):
        self.base_url = base_url
        self.worker_interval_minutes = worker_interval_minutes
        self.initial_top_10 = []
        self.test_users = []
        self.submitted_scores = {}
        
    def fetch_top_leaderboard(self) -> List[Dict]:
        """Fetch top 10 leaderboard"""
        try:
            response = requests.get(f"{self.base_url}/top", timeout=10)
            response.raise_for_status()
            data = response.json()
            
            if data.get("success"):
                return data.get("data", [])
            else:
                print(f"❌ Failed to fetch top leaderboard: {data}")
                return []
        except Exception as e:
            print(f"❌ Error fetching top leaderboard: {e}")
            return []
    
    def fetch_user_rank(self, user_id: int) -> Dict:
        """Fetch rank for a specific user"""
        try:
            response = requests.get(f"{self.base_url}/rank/{user_id}", timeout=10)
            response.raise_for_status()
            data = response.json()
            
            if data.get("success"):
                return data.get("data", {})
            else:
                print(f"❌ Failed to fetch user {user_id} rank: {data}")
                return {}
        except Exception as e:
            print(f"❌ Error fetching user {user_id} rank: {e}")
            return {}
    
    def submit_score(self, user_id: int, score: int, game_mode: str = "solo") -> bool:
        """Submit a score for a user"""
        try:
            payload = {
                "user_id": user_id,
                "score": score,
                "game_mode": game_mode  # Added required field
            }
            
            response = requests.post(
                f"{self.base_url}/submit",
                json=payload,
                timeout=10
            )
            response.raise_for_status()
            
            # Track submitted scores
            if user_id not in self.submitted_scores:
                self.submitted_scores[user_id] = []
            self.submitted_scores[user_id].append(score)
            
            print(f"✅ Submitted score: user_id={user_id}, score={score}, game_mode={game_mode}")
            return True
        except requests.exceptions.HTTPError as e:
            print(f"❌ HTTP Error submitting score for user {user_id}: {e}")
            if e.response is not None:
                print(f"   Response Body: {e.response.text}")
            return False
        except Exception as e:
            print(f"❌ Error submitting score for user {user_id}: {e}")
            return False
    
    def print_leaderboard(self, leaderboard: List[Dict], title: str):
        """Print leaderboard in a nice format"""
        print(f"\n{'='*80}")
        print(f"{title}")
        print(f"{'='*80}")
        print(f"{'Rank':<6} {'User ID':<10} {'Total Score':<12} {'Leaderboard ID':<15}")
        print(f"{'-'*80}")
        
        for entry in leaderboard:
            print(f"{entry['rank']:<6} {entry['user_id']:<10} {entry['total_score']:<12} {entry['id']:<15}")
        print(f"{'='*80}\n")
    
    def select_random_users_outside_top10(self, count: int = 5) -> List[int]:
        """Select random users NOT in top 10"""
        # Avoid users in top 10
        top_10_user_ids = {entry['user_id'] for entry in self.initial_top_10}
        
        # Generate random user IDs between 1 and 1,000,000, excluding top 10
        test_users = []
        attempts = 0
        max_attempts = count * 10
        
        while len(test_users) < count and attempts < max_attempts:
            user_id = random.randint(10000, 100000)  # Mid-range users
            if user_id not in top_10_user_ids and user_id not in test_users:
                test_users.append(user_id)
            attempts += 1
        
        return test_users
    
    def calculate_expected_rank(self, user_id: int, new_total_score: int, 
                                 current_leaderboard: List[Dict]) -> int:
        """Calculate expected rank based on new total score"""
        # Count how many users have higher scores
        higher_scores = 0
        for entry in current_leaderboard:
            if entry['user_id'] == user_id:
                continue  # Skip self
            if entry['total_score'] > new_total_score:
                higher_scores += 1
        
        return higher_scores + 1
    
    def wait_for_batch_processing(self):
        """Wait for batch worker to complete processing"""
        # Wait for worker interval + 30 second buffer
        wait_time = (self.worker_interval_minutes * 60) + 30
        
        print(f"\n⏳ Step 5: Waiting for batch recalculation...")
        print(f"Worker runs every {self.worker_interval_minutes} minutes")
        print(f"Waiting {wait_time} seconds ({self.worker_interval_minutes} min + 30s buffer)\n")
        
        # Countdown timer
        start_time = time.time()
        while True:
            elapsed = int(time.time() - start_time)
            remaining = wait_time - elapsed
            
            if remaining <= 0:
                break
            
            mins, secs = divmod(remaining, 60)
            print(f"⏱️  Time remaining: {mins:02d}:{secs:02d} ({elapsed}s elapsed)", end='\r')
            time.sleep(1)
        
        print(f"\n✅ Wait complete! Total wait time: {wait_time} seconds\n")
    
    def run_test(self):
        """Main test workflow"""
        print("\n" + "="*80)
        print("🚀 STARTING LEADERBOARD CORRECTNESS TEST")
        print(f"⚙️  Worker Interval: {self.worker_interval_minutes} minutes")
        print("="*80 + "\n")
        
        # Step 1: Fetch initial top 10
        print("📊 Step 1: Fetching initial top 10 leaderboard...")
        self.initial_top_10 = self.fetch_top_leaderboard()
        
        if not self.initial_top_10:
            print("❌ Failed to fetch initial leaderboard. Exiting.")
            return
        
        self.print_leaderboard(self.initial_top_10, "INITIAL TOP 10 LEADERBOARD")
        
        # Step 2: Select random users outside top 10
        print("🎲 Step 2: Selecting random test users (not in top 10)...")
        self.test_users = self.select_random_users_outside_top10(count=5)
        print(f"Selected test users: {self.test_users}\n")
        
        # Step 3: Fetch current rank for each test user
        print("📈 Step 3: Fetching current ranks for test users...")
        initial_user_data = {}
        
        for user_id in self.test_users:
            user_data = self.fetch_user_rank(user_id)
            if user_data:
                initial_user_data[user_id] = user_data
                print(f"User {user_id}: Rank={user_data.get('rank')}, "
                      f"Total Score={user_data.get('total_score')}")
            time.sleep(0.1)
        
        print(f"\n✅ Fetched data for {len(initial_user_data)} users\n")
        
        # Step 4: Submit high scores for test users
        print("🎯 Step 4: Submitting HIGH scores for test users...")
        print("(These scores should push users into top 10)\n")
        
        high_scores = [9000, 9500, 8800, 9200, 8900]  # Scores likely to be in top 10
        game_modes = ["solo", "team"]  # Mix of game modes
        
        submission_time = datetime.now()
        print(f"⏰ Submission start time: {submission_time.strftime('%H:%M:%S')}\n")
        
        for i, user_id in enumerate(self.test_users):
            score = high_scores[i % len(high_scores)]
            game_mode = random.choice(game_modes)  # Randomly choose solo or team
            self.submit_score(user_id, score, game_mode)
            time.sleep(0.2)  # Small delay between submissions
        
        submission_end_time = datetime.now()
        print(f"\n✅ Submitted {len(self.test_users)} high scores")
        print(f"⏰ Submission end time: {submission_end_time.strftime('%H:%M:%S')}")
        
        # Step 5: Wait for batch processing
        self.wait_for_batch_processing()
        
        # Step 6: Fetch new top 10 leaderboard
        print("📊 Step 6: Fetching NEW top 10 leaderboard...")
        verification_time = datetime.now()
        print(f"⏰ Verification time: {verification_time.strftime('%H:%M:%S')}\n")
        
        new_top_10 = self.fetch_top_leaderboard()
        
        if not new_top_10:
            print("❌ Failed to fetch new leaderboard. Exiting.")
            return
        
        self.print_leaderboard(new_top_10, "NEW TOP 10 LEADERBOARD (AFTER UPDATES)")
        
        # Step 7: Verify test users are now in top 10
        print("🔍 Step 7: Verifying test users' new ranks...\n")
        
        verification_results = []
        new_top_10_user_ids = {entry['user_id'] for entry in new_top_10}
        
        for user_id in self.test_users:
            new_user_data = self.fetch_user_rank(user_id)
            
            if not new_user_data:
                print(f"❌ User {user_id}: Failed to fetch new data")
                continue
            
            old_total = initial_user_data.get(user_id, {}).get('total_score', 0)
            new_total = new_user_data.get('total_score', 0)
            new_rank = new_user_data.get('rank', 0)
            
            # Calculate expected total score
            submitted = sum(self.submitted_scores.get(user_id, []))
            expected_total = old_total + submitted
            
            # Check if in top 10
            in_top_10 = user_id in new_top_10_user_ids
            
            # Determine if correct
            score_correct = (new_total == expected_total)
            rank_plausible = (new_rank <= 10 if in_top_10 else new_rank > 10)
            
            result = {
                'user_id': user_id,
                'old_total': old_total,
                'submitted': submitted,
                'expected_total': expected_total,
                'actual_total': new_total,
                'new_rank': new_rank,
                'in_top_10': in_top_10,
                'score_correct': score_correct,
                'rank_plausible': rank_plausible,
                'overall_correct': score_correct and rank_plausible
            }
            
            verification_results.append(result)
            time.sleep(0.1)
        
        # Step 8: Print verification results
        print("\n" + "="*80)
        print("📋 VERIFICATION RESULTS")
        print("="*80 + "\n")
        
        for result in verification_results:
            status = "✅ PASS" if result['overall_correct'] else "❌ FAIL"
            print(f"{status} User {result['user_id']}:")
            print(f"  Old Total Score:      {result['old_total']}")
            print(f"  Submitted Score:      {result['submitted']}")
            print(f"  Expected Total Score: {result['expected_total']}")
            print(f"  Actual Total Score:   {result['actual_total']} "
                  f"{'✅' if result['score_correct'] else '❌ MISMATCH!'}")
            print(f"  New Rank:             {result['new_rank']}")
            print(f"  In Top 10:            {result['in_top_10']}")
            print()
        
        # Step 9: Summary
        print("="*80)
        print("📊 TEST SUMMARY")
        print("="*80)
        
        passed = sum(1 for r in verification_results if r['overall_correct'])
        total = len(verification_results)
        
        print(f"Total Tests:  {total}")
        print(f"Passed:       {passed} ✅")
        print(f"Failed:       {total - passed} ❌")
        print(f"Success Rate: {(passed/total*100):.1f}%")
        
        # Calculate timing
        total_test_time = (verification_time - submission_time).total_seconds()
        print(f"\nTest Duration: {int(total_test_time)} seconds ({total_test_time/60:.1f} minutes)")
        
        if passed == total:
            print("\n🎉 ALL TESTS PASSED! Your leaderboard system is working correctly!")
        else:
            print("\n⚠️  SOME TESTS FAILED. Check the verification results above.")
        
        print("="*80 + "\n")


if __name__ == "__main__":
    # Set worker_interval_minutes to match your Go code (3 minutes)
    tester = LeaderboardTester(
        base_url=API_BASE_URL,
        worker_interval_minutes=3  # Your worker runs every 3 minutes
    )
    tester.run_test()