package router

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"

	beaconv1 "github.com/gulugulu3399/bifrost/api/content/v1/beacon"
	nexusv1 "github.com/gulugulu3399/bifrost/api/content/v1/nexus"
	"github.com/gulugulu3399/bifrost/internal/pkg/observability/logger"
)

type Router struct {
	nexus  nexusv1.PostServiceClient
	beacon beaconv1.BeaconServiceClient

	marshal protojson.MarshalOptions
}

// New 创建 HTTP Router（不依赖 grpc-gateway 生成代码）
// 说明：本仓库目前没有生成 Register*Handler 的 gateway 文件，因此这里用“手写薄适配层”方式
// 直接把 HTTP 请求转为 gRPC 调用。
func New(_ context.Context, nexusConn *grpc.ClientConn, beaconConn *grpc.ClientConn) (http.Handler, error) {
	r := &Router{
		nexus:  nexusv1.NewPostServiceClient(nexusConn),
		beacon: beaconv1.NewBeaconServiceClient(beaconConn),
		marshal: protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		},
	}

	mux := http.NewServeMux()

	// Read API (Beacon)
	mux.HandleFunc("/v1/posts", r.handleListPosts)
	mux.HandleFunc("/v1/posts/", r.handleGetPost) // /v1/posts/{id}

	// Write API (Nexus)
	mux.HandleFunc("/v1/admin/posts", r.handleCreatePost)
	mux.HandleFunc("/v1/admin/posts/", r.handleUpdateOrDeletePost) // /v1/admin/posts/{id}

	return mux, nil
}

func writeJSON(w http.ResponseWriter, status int, b []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(b)
}

func (r *Router) marshalResp(w http.ResponseWriter, status int, msg interface{ ProtoReflect() }) {
	b, err := r.marshal.Marshal(msg)
	if err != nil {
		logger.Global().Error("marshal response failed", logger.Err(err))
		writeJSON(w, http.StatusInternalServerError, []byte(`{"code":500,"message":"marshal failed"}`))
		return
	}
	writeJSON(w, status, b)
}

// ---------- Beacon handlers ----------

func (r *Router) handleListPosts(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	q := req.URL.Query()
	page, _ := strconv.ParseInt(q.Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(q.Get("page_size"), 10, 32)

	resp, err := r.beacon.ListPosts(req.Context(), &beaconv1.ListPostsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Keyword:  q.Get("keyword"),
	})
	if err != nil {
		logger.Global().Warn("beacon ListPosts failed", logger.Err(err))
		writeJSON(w, http.StatusBadGateway, []byte(`{"code":502,"message":"upstream error"}`))
		return
	}

	r.marshalResp(w, http.StatusOK, resp)
}

func (r *Router) handleGetPost(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	idStr := req.URL.Path[len("/v1/posts/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, []byte(`{"code":400,"message":"invalid id"}`))
		return
	}

	resp, err := r.beacon.GetPost(req.Context(), &beaconv1.GetPostRequest{PostId: id})
	if err != nil {
		logger.Global().Warn("beacon GetPost failed", logger.Err(err))
		writeJSON(w, http.StatusBadGateway, []byte(`{"code":502,"message":"upstream error"}`))
		return
	}
	r.marshalResp(w, http.StatusOK, resp)
}

// ---------- Nexus handlers ----------

type rawBody struct {
	b []byte
}

func readBody(req *http.Request) ([]byte, bool) {
	b, err := io.ReadAll(req.Body)
	_ = req.Body.Close()
	if err != nil {
		return nil, false
	}
	return b, true
}

func (r *Router) handleCreatePost(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	b, ok := readBody(req)
	if !ok {
		writeJSON(w, http.StatusBadRequest, []byte(`{"code":400,"message":"read body failed"}`))
		return
	}

	in := &nexusv1.CreatePostRequest{}
	if err := protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(b, in); err != nil {
		writeJSON(w, http.StatusBadRequest, []byte(`{"code":400,"message":"invalid json"}`))
		return
	}

	resp, err := r.nexus.CreatePost(req.Context(), in)
	if err != nil {
		logger.Global().Warn("nexus CreatePost failed", logger.Err(err))
		writeJSON(w, http.StatusBadGateway, []byte(`{"code":502,"message":"upstream error"}`))
		return
	}
	r.marshalResp(w, http.StatusOK, resp)
}

func (r *Router) handleUpdateOrDeletePost(w http.ResponseWriter, req *http.Request) {
	idStr := req.URL.Path[len("/v1/admin/posts/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, []byte(`{"code":400,"message":"invalid id"}`))
		return
	}

	switch req.Method {
	case http.MethodPut:
		b, ok := readBody(req)
		if !ok {
			writeJSON(w, http.StatusBadRequest, []byte(`{"code":400,"message":"read body failed"}`))
			return
		}
		in := &nexusv1.UpdatePostRequest{}
		if err := protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(b, in); err != nil {
			writeJSON(w, http.StatusBadRequest, []byte(`{"code":400,"message":"invalid json"}`))
			return
		}
		in.PostId = id

		resp, err := r.nexus.UpdatePost(req.Context(), in)
		if err != nil {
			logger.Global().Warn("nexus UpdatePost failed", logger.Err(err))
			writeJSON(w, http.StatusBadGateway, []byte(`{"code":502,"message":"upstream error"}`))
			return
		}
		r.marshalResp(w, http.StatusOK, resp)
	case http.MethodDelete:
		resp, err := r.nexus.DeletePost(req.Context(), &nexusv1.DeletePostRequest{PostId: id})
		if err != nil {
			logger.Global().Warn("nexus DeletePost failed", logger.Err(err))
			writeJSON(w, http.StatusBadGateway, []byte(`{"code":502,"message":"upstream error"}`))
			return
		}
		r.marshalResp(w, http.StatusOK, resp)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

