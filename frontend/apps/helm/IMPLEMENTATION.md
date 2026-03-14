# Helm - Bifrost Admin Dashboard

## Architecture

Helm follows the **Tauri Sidecar Pattern**:

- **Frontend (Vue 3)**: UI rendering and user interaction only
- **Backend (Rust)**: ALL business logic, data processing, networking, and file manipulation

## Features Implemented

### 1. Network Layer (HTTP/REST)

- ✅ HTTP client using `reqwest`
- ✅ Authentication with JWT tokens
- ✅ Communication with Gjallar Gateway at `http://localhost:8080`

### 2. Smart Image Upload

- ✅ Image selection from local filesystem
- ✅ Automatic WebP conversion with 75% quality
- ✅ Intelligent resizing (max 1920px)
- ✅ Multipart upload to backend storage

### 3. Tauri Commands Exposed

#### `login_cmd(username, password)`

Authenticates user and stores JWT token in Rust memory.

**Usage:**

```typescript
const response = await invoke<LoginResponse>("login_cmd", {
  username: "admin",
  password: "password123",
});
```

#### `upload_image_cmd(file_path)`

Processes image to WebP and uploads to server.

**Usage:**

```typescript
const response = await invoke<UploadResponse>("upload_image_cmd", {
  filePath: "/path/to/image.jpg",
});
```

#### `is_authenticated()`

Checks if user is currently authenticated.

**Usage:**

```typescript
const isLoggedIn = await invoke<boolean>("is_authenticated");
```

#### `logout_cmd()`

Clears the stored authentication token.

**Usage:**

```typescript
await invoke("logout_cmd");
```

## Development

### Prerequisites

- Rust 1.75+
- Node.js 18+
- pnpm

### Install Dependencies

```bash
# Install Rust dependencies
cd src-tauri
cargo build

# Install Node dependencies
cd ..
pnpm install
```

### Run Development Server

```bash
pnpm tauri dev
```

### Build for Production

```bash
pnpm tauri build
```

## Environment Variables

Set `GJALLAR_URL` to override the default backend URL:

```bash
export GJALLAR_URL=http://localhost:8080
```

## API Endpoints Used

- `POST /v1/auth/login` - User authentication
- `POST /v1/storage/upload` - Image upload (multipart/form-data)

## Rust Modules

### `http_client.rs`

Handles all HTTP communication with the backend:

- Login/authentication
- Token management
- File uploads

### `image_processor.rs`

Handles image processing:

- WebP conversion
- Intelligent resizing
- Quality optimization

### `lib.rs`

Main entry point:

- Tauri command handlers
- State management
- Application lifecycle

## Type Safety

All commands use strongly typed inputs and outputs:

```rust
// Input types
LoginRequest { username: String, password: String }

// Output types
LoginResponse { token: String, user_id: Option<String>, nickname: Option<String> }
UploadResponse { key: String, url: Option<String> }
```

## Error Handling

All commands return `Result<T, String>` for proper error handling in TypeScript:

```typescript
try {
  const result = await invoke("upload_image_cmd", { filePath });
} catch (error) {
  console.error("Upload failed:", error);
}
```

## Performance Considerations

- Image processing runs in blocking thread pool to avoid blocking main thread
- HTTP requests are async and non-blocking
- Token stored in Arc<RwLock> for thread-safe access

## Future Enhancements

- [ ] Add gRPC support (migrate from HTTP)
- [ ] Implement better WebP quality control with `webp` crate
- [ ] Add progress callbacks for large file uploads
- [ ] Implement image thumbnail generation
- [ ] Add batch upload support
