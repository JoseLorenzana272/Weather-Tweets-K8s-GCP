use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct WeatherTweet {
    pub description: String,
    pub country: String,
    pub weather: String,
}