use anyhow::{Context, Result};
use image::GenericImageView;
use image::{imageops::FilterType, DynamicImage, ImageFormat};
use std::io::Cursor;

/// Image processing utilities
pub struct ImageProcessor;

impl ImageProcessor {
    const MAX_WIDTH: u32 = 1920;
    const MAX_HEIGHT: u32 = 1920;
    const WEBP_QUALITY: u8 = 75;

    /// Process an image: resize if needed and convert to WebP
    pub fn process_to_webp(image_data: Vec<u8>) -> Result<Vec<u8>> {
        // Load the image
        let img = image::load_from_memory(&image_data).context("Failed to decode image")?;

        // Resize if necessary
        let processed_img = Self::resize_if_needed(img);

        // Encode to WebP
        Self::encode_webp(processed_img)
    }

    /// Process an image from a file path
    pub fn process_file_to_webp(file_path: &str) -> Result<Vec<u8>> {
        // Read the file
        let img =
            image::open(file_path).context(format!("Failed to open image file: {}", file_path))?;

        // Resize if necessary
        let processed_img = Self::resize_if_needed(img);

        // Encode to WebP
        Self::encode_webp(processed_img)
    }

    /// Resize image if it exceeds max dimensions
    fn resize_if_needed(img: DynamicImage) -> DynamicImage {
        let (width, height) = img.dimensions();

        if width > Self::MAX_WIDTH || height > Self::MAX_HEIGHT {
            // Calculate new dimensions while maintaining aspect ratio
            let ratio = (width as f32 / Self::MAX_WIDTH as f32)
                .max(height as f32 / Self::MAX_HEIGHT as f32);

            let new_width = (width as f32 / ratio) as u32;
            let new_height = (height as f32 / ratio) as u32;

            img.resize(new_width, new_height, FilterType::Lanczos3)
        } else {
            img
        }
    }

    /// Encode image to WebP format
    fn encode_webp(img: DynamicImage) -> Result<Vec<u8>> {
        let mut buffer = Cursor::new(Vec::new());

        // For lossy encoding with quality control, we need to use the webp crate directly.
        // For now, use image crate to write WebP as a fallback.
        img.write_to(&mut buffer, ImageFormat::WebP)
            .context("Failed to encode image to WebP")?;

        Ok(buffer.into_inner())
    }

    /// Get the original file extension for naming
    pub fn get_extension(file_path: &str) -> String {
        std::path::Path::new(file_path)
            .extension()
            .and_then(|ext| ext.to_str())
            .unwrap_or("jpg")
            .to_string()
    }

    /// Generate a WebP filename from the original path
    pub fn generate_webp_filename(original_path: &str) -> String {
        let stem = std::path::Path::new(original_path)
            .file_stem()
            .and_then(|s| s.to_str())
            .unwrap_or("image");

        format!("{}.webp", stem)
    }
}
