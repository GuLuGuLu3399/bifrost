#!/usr/bin/env bash

# gRPC Gateway 集成测试脚本
# 说明：本脚本假设已启动 beacon 和 nexus 微服务，以及 gjallar 网关

set -e

BASE_URL="http://localhost:8080"
BEACON_PORT=50051
NEXUS_PORT=50052

echo "🧪 gRPC Gateway 集成测试"
echo "=================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试函数
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected_code=$4
    
    echo -n "📌 $method $endpoint ... "
    
    if [ -z "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi
    
    http_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "$expected_code" ]; then
        echo -e "${GREEN}✅ HTTP $http_code${NC}"
        echo "   响应: $body" | head -c 100
        echo ""
    else
        echo -e "${RED}❌ HTTP $http_code (期望 $expected_code)${NC}"
        echo "   响应: $body"
        return 1
    fi
}

# 检查网关连接
echo "📡 检查网关连接..."
if ! curl -s "$BASE_URL/v1/posts" > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  网关未响应，请确保已启动服务：${NC}"
    echo "   1. 启动 Beacon 服务:  cd go_services && go run ./cmd/beacon/main.go"
    echo "   2. 启动 Nexus 服务:   cd go_services && go run ./cmd/nexus/main.go"
    echo "   3. 启动 Gjallar 网关: cd go_services && go run ./cmd/gjallar/main.go"
    exit 1
fi
echo -e "${GREEN}✅ 网关连接正常${NC}"
echo ""

# Beacon 服务测试 (读服务)
echo "📖 Beacon 服务测试（读服务）"
echo "---"

test_endpoint "GET" "/v1/posts" "" 200
test_endpoint "GET" "/v1/posts?page_size=10" "" 200
test_endpoint "GET" "/v1/posts/1" "" 200

test_endpoint "GET" "/v1/categories" "" 200
test_endpoint "GET" "/v1/tags" "" 200

echo ""

# Nexus 服务测试 (写服务)
echo "✍️  Nexus 服务测试（写服务）"
echo "---"

# 注册
test_endpoint "POST" "/v1/auth/register" '{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123",
  "nickname": "Test User"
}' 200

# 登录
test_endpoint "POST" "/v1/auth/login" '{
  "identifier": "testuser",
  "password": "password123"
}' 200

# 创建文章
test_endpoint "POST" "/v1/posts" '{
  "title": "Test Post",
  "slug": "test-post",
  "raw_markdown": "# Hello World",
  "category_id": 1,
  "tag_names": ["test", "demo"],
  "status": 1,
  "cover_image_key": "cover.jpg"
}' 200

echo ""
echo "=================================="
echo -e "${GREEN}✅ 测试完成！${NC}"
echo ""
echo "💡 验证网关正常工作："
echo "   - gRPC Gateway 自动将 HTTP 请求转换为 gRPC 调用"
echo "   - 响应自动从 Protobuf 转换为 JSON"
echo "   - URL 路径参数自动提取和映射"
echo ""
