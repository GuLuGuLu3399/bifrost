use std::path::PathBuf;
use std::{env, fs};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Docker 构建时 crate 根在 /workspace/common，API 在 /workspace/api（相对 ../api）
    // 本地构建时 crate 根在 rust_services/common，API 在 ../../api
    let api_root = if std::path::Path::new("../api").exists() {
        "../api" // Docker 构建路径
    } else {
        "../../api" // 本地构建路径
    };

    let protos: Vec<String> = vec![
        format!("{}/common/v1/common.proto", api_root),
        format!("{}/content/v1/models.proto", api_root),
        format!("{}/content/v1/forge/forge.proto", api_root),
        format!("{}/content/v1/oracle/oracle.proto", api_root),
        format!("{}/search/v1/mirror.proto", api_root),
    ];

    println!("cargo:rerun-if-changed={}", api_root);
    for p in &protos {
        println!("cargo:rerun-if-changed={}", p);
    }

    let out_dir = PathBuf::from(env::var("OUT_DIR")?);
    fs::create_dir_all(&out_dir)?; // 确保 OUT_DIR 存在

    let descriptor_path = out_dir.join("bifrost_descriptor.bin");
    let proto_refs: Vec<&str> = protos.iter().map(|s| s.as_str()).collect();

    // include path 指向 api 的父目录、api 根，以及可选的第三方 googleapis 目录
    let api_parent = std::path::Path::new(api_root)
        .parent()
        .unwrap_or_else(|| std::path::Path::new("."));
    let api_parent_str = api_parent.to_str().unwrap_or(".");

    // 检测第三方 googleapis 目录（Docker 构建时为 ../third_party/googleapis，本地为 ../../third_party/googleapis）
    let third_party_root = if std::path::Path::new("../third_party/googleapis").exists() {
        "../third_party/googleapis"
    } else if std::path::Path::new("../../third_party/googleapis").exists() {
        "../../third_party/googleapis"
    } else {
        ""
    };

    // Always include common/proto so vendored google/api/*.proto can satisfy imports.
    let mut include_paths = vec![api_parent_str, api_root, "proto"];
    if !third_party_root.is_empty() {
        include_paths.push(third_party_root);
    }

    tonic_prost_build::configure()
        .build_server(true)
        .build_client(true)
        .file_descriptor_set_path(&descriptor_path)
        .out_dir(&out_dir)
        .compile_protos(&proto_refs, &include_paths)?;

    let descriptor_bytes = fs::read(&descriptor_path)?;
    pbjson_build::Builder::new()
        .register_descriptors(&descriptor_bytes)?
        .build(&[".bifrost"])?;

    Ok(())
}
