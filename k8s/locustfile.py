from locust import HttpUser, task, between
import random

class WeatherUser(HttpUser):
    wait_time = between(1, 3)
    host = "http://34.69.223.128.nip.io"

    @task
    def send_weather_message(self):
        country = random.choice(["GT", "ESP", "SWN"])
        description = f"Test message from {country}"
        weather = random.choice(["Sunny", "Cloudy", "Rainy"])
        self.client.post(
            "/input",
            json={"description": description, "country": country, "weather": weather},
            headers={"Content-Type": "application/json"}
        )