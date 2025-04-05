use actix_web::{post, web, App, HttpServer, Responder, HttpResponse};
use serde::{Deserialize, Serialize};
use std::fmt::Debug;

#[derive(Debug, Serialize, Deserialize)]
struct WeatherTweet {
    description: String,
    country: String,
    weather: String,
}

#[post("/input")]
async fn process_tweet(tweet: web::Json<WeatherTweet>) -> impl Responder {
    println!("Received tweet: {:?}", tweet);
    
    HttpResponse::Ok().json(serde_json::json!({
        "status": "received",
        "country": tweet.country,
        "weather": tweet.weather
    }))
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    println!("Starting Rust weather API on http://localhost:8080");
    
    HttpServer::new(|| {
        App::new()
            .service(process_tweet)
    })
    .bind("0.0.0.0:8080")?
    .run()
    .await
}
