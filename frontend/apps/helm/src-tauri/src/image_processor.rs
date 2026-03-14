use anyhow::{Context, Result};
use image::GenericImageView;
use image::{imageops::FilterType, DynamicImage, ImageFormat};
use std::io::Cursor;

/// 图片处理工具
pub struct ImageProcessor;

impl ImageProcessor {
    const MAX_WIDTH: u32 = 1920;
    const MAX_HEIGHT: u32 = 1920;
    const WEBP_QUALITY: u8 = 75;

    /// 处理图片：按需缩放并转换为 WebP
    pub fn process_to_webp(image_data: Vec<u8>) -> Result<Vec<u8>> {
        // 读取图片二进制并解码
        let img = image::load_from_memory(&image_data).context("Failed to decode image")?;

        // 超出限制时按比例缩放
        let processed_img = Self::resize_if_needed(img);

        // 编码为 WebP
        Self::encode_webp(processed_img)
    }

    /// 从文件路径读取并处理图片为 WebP
    pub fn process_file_to_webp(file_path: &str) -> Result<Vec<u8>> {
        // 从磁盘打开图片文件
        let img =
            image::open(file_path).context(format!("Failed to open image file: {}", file_path))?;

        // 超出限制时按比例缩放
        let processed_img = Self::resize_if_needed(img);

        // 编码为 WebP
        Self::encode_webp(processed_img)
    }

    /// 当图片超过最大尺寸时进行等比缩放
    fn resize_if_needed(img: DynamicImage) -> DynamicImage {
        let (width, height) = img.dimensions();

        if width > Self::MAX_WIDTH || height > Self::MAX_HEIGHT {
            // 在保持宽高比的前提下计算新尺寸
            let ratio = (width as f32 / Self::MAX_WIDTH as f32)
                .max(height as f32 / Self::MAX_HEIGHT as f32);

            let new_width = (width as f32 / ratio) as u32;
            let new_height = (height as f32 / ratio) as u32;

            img.resize(new_width, new_height, FilterType::Lanczos3)
        } else {
            img
        }
    }

    /// 将图片编码为 WebP
    fn encode_webp(img: DynamicImage) -> Result<Vec<u8>> {
        let mut buffer = Cursor::new(Vec::new());

        // 如果需要可控质量的有损编码，可考虑引入 webp crate。
        // 当前先使用 image crate 作为 WebP 编码兜底方案。
        img.write_to(&mut buffer, ImageFormat::WebP)
            .context("Failed to encode image to WebP")?;

        Ok(buffer.into_inner())
    }

    /// 获取原始扩展名（用于命名）
    pub fn get_extension(file_path: &str) -> String {
        std::path::Path::new(file_path)
            .extension()
            .and_then(|ext| ext.to_str())
            .unwrap_or("jpg")
            .to_string()
    }

    /// 基于原始路径生成 WebP 文件名
    pub fn generate_webp_filename(original_path: &str) -> String {
        let stem = std::path::Path::new(original_path)
            .file_stem()
            .and_then(|s| s.to_str())
            .unwrap_or("image");

        format!("{}.webp", stem)
    }
}
