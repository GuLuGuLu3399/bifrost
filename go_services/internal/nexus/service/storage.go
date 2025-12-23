package service

import (
    "context"
    "fmt"
    "path/filepath"
    "time"

    nexusv1 "github.com/gulugulu3399/bifrost/api/content/v1/nexus"
    pkgcfg "github.com/gulugulu3399/bifrost/internal/pkg/config"
    "github.com/gulugulu3399/bifrost/internal/pkg/contextx"
    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type StorageService struct {
    nexusv1.UnimplementedStorageServiceServer
    client *minio.Client
    bucket string
}

// NewStorageService 初始化 MinIO 客户端
func NewStorageService(cfg *pkgcfg.NexusConfig) (*StorageService, error) {
    m := cfg.Storage
    cli, err := minio.New(m.Endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(m.AccessKeyID, m.SecretAccessKey, ""),
        Secure: m.UseSSL,
        Region: m.Region,
    })
    if err != nil {
        return nil, err
    }
    return &StorageService{client: cli, bucket: m.Bucket}, nil
}

// GetUploadTicket 生成预签名上传 URL
func (s *StorageService) GetUploadTicket(ctx context.Context, req *nexusv1.GetUploadTicketRequest) (*nexusv1.GetUploadTicketResponse, error) {
    uid := contextx.UserIDFromContext(ctx)
    if uid == 0 {
        return nil, status.Error(codes.Unauthenticated, "authentication required")
    }

    filename := req.GetFilename()
    if filename == "" {
        filename = "file.bin"
    }
    ext := filepath.Ext(filename)
    if ext == "" {
        ext = ".bin"
    }

    // 依据用途分类存储路径
    var prefix string
    switch req.GetUsage() {
    case "avatar":
        prefix = fmt.Sprintf("avatars/%d", uid)
    case "cover":
        prefix = fmt.Sprintf("covers/%d", uid)
    case "post_image":
        prefix = fmt.Sprintf("posts/%d", uid)
    default:
        return nil, status.Error(codes.InvalidArgument, "invalid usage type")
    }

    objectKey := fmt.Sprintf("%s/%d_%d%s", prefix, uid, time.Now().UnixNano(), ext)
    expiry := 10 * time.Minute

    url, err := s.client.PresignedPutObject(ctx, s.bucket, objectKey, expiry)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "generate presigned url failed: %v", err)
    }

    return &nexusv1.GetUploadTicketResponse{UploadUrl: url.String(), ObjectKey: objectKey}, nil
}
