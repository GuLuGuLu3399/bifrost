use std::{env, fs};
use std::path::PathBuf;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let api_root = "../../api";
    let repo_root = "../..";

    let protos = &[
        "../../api/common/v1/common.proto",
        "../../api/content/v1/models.proto",
        "../../api/content/v1/forge/forge.proto",
        "../../api/content/v1/oracle/oracle.proto",
        "../../api/search/v1/mirror.proto",
    ];

    println!("cargo:rerun-if-changed={}", api_root);
    for p in protos {
        println!("cargo:rerun-if-changed={}", p);
    }

    let out_dir = PathBuf::from(env::var("OUT_DIR")?);
    fs::create_dir_all(&out_dir)?; // 确保 OUT_DIR 存在

    let descriptor_path = out_dir.join("bifrost_descriptor.bin");

    tonic_prost_build::configure()
        .build_server(true)
        .build_client(true)
        .file_descriptor_set_path(&descriptor_path)
        .out_dir(&out_dir)
        .compile_protos(protos, &[repo_root, api_root])?;

    let descriptor_bytes = fs::read(&descriptor_path)?;
    pbjson_build::Builder::new()
        .register_descriptors(&descriptor_bytes)?
        .build(&[".bifrost"])?;

    Ok(())
}
