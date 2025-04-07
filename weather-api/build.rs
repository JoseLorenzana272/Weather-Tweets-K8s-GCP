use std::path::PathBuf;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let proto_file = "../proto/weather.proto";
    let proto_path = "../proto";
    
    if !PathBuf::from(proto_file).exists() {
        panic!("Proto file not found at: {}", proto_file);
    }

    tonic_build::configure()
        .build_server(false)
        .compile(&[proto_file], &[proto_path])?;
    
    println!("Proto file compiled successfully");
    Ok(())
}