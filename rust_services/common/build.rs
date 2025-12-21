fn main() -> Result<(), Box<dyn std::error::Error>> {
    // 1. 定义 Proto 根目录
    // 结构: rust_services/common/build.rs -> ../../api
    let proto_root = "../../api";

    // 2. 定义需要编译的具体文件
    // 注意：一定要包含所有被引用的文件 (如 models.proto)
    let protos = &[
        "../../api/common/v1/common.proto",
        "../../api/content/v1/models.proto",
        "../../api/content/v1/nexus.proto",  // Nexus 写服务接口
        "../../api/content/v1/beacon.proto", // Beacon 读服务接口
        "../../api/search/v1/mirror.proto",  // Mirror 搜索接口
    ];

    // 3. 监控文件变动 (当 proto 改变时自动重编)
    println!("cargo:rerun-if-changed={}", proto_root);

    // 4. 执行编译
    tonic_build::configure()
        .build_server(true)  // 生成 Server 端代码 (Trait)
        .build_client(true)  // 生成 Client 端代码
        .out_dir("src/generated") // (可选) 显式指定输出目录，或者默认在 OUT_DIR
        .compile(protos, &[proto_root])?; // 第二个参数是 include 路径

    Ok(())
}