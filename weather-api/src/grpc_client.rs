use tonic::{transport::Channel, Request};

pub mod weather {
    tonic::include_proto!("weather");
}

pub async fn send_to_go_grpc(
    tweet: weather::WeatherTweet,
) -> Result<(), tonic::Status> {
    let channel = Channel::from_static("http://go-api:50051")
        .connect()
        .await
        .map_err(|e| tonic::Status::new(tonic::Code::Internal, e.to_string()))?;

    let mut client = weather::weather_service_client::WeatherServiceClient::new(channel);
    let request = Request::new(tweet);

    match client.process_tweet(request).await {
        Ok(response) => {
            println!("Go gRPC response: {:?}", response.into_inner());
            Ok(())
        },
        Err(e) => {
            eprintln!("gRPC call failed: {:?}", e);
            Err(e)
        }
    }
}