mod grpc_client;
mod models;

use actix_web::{get, post, web, App, HttpResponse, HttpServer};
use models::WeatherTweet;

#[get("/health")]
async fn health_check() -> HttpResponse {
    HttpResponse::Ok().body("OK")
}

#[post("/input")]
async fn handle_weather_tweet(tweet: web::Json<WeatherTweet>) -> HttpResponse {
    println!("Rust received: {:?}", tweet);

    // Convert to gRPC message
    let grpc_tweet = grpc_client::weather::WeatherTweet {
        description: tweet.description.clone(),
        country: tweet.country.clone(),
        weather: tweet.weather.clone(),
    };

    // Call Go gRPC server
    match grpc_client::send_to_go_grpc(grpc_tweet).await {
        Ok(_) => HttpResponse::Ok().json(tweet),
        Err(e) => {
            eprintln!("gRPC error: {:?}", e);
            HttpResponse::InternalServerError().finish()
        }
    }
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    HttpServer::new(|| {
        App::new()
            .service(health_check)
            .service(handle_weather_tweet)
    })
    .bind("0.0.0.0:8080")?
    .run()
    .await
}