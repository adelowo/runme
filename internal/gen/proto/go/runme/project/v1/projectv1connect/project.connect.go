// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: runme/project/v1/project.proto

package projectv1connect

import (
	context "context"
	errors "errors"
	connect_go "github.com/bufbuild/connect-go"
	v1 "github.com/stateful/runme/internal/gen/proto/go/runme/project/v1"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect_go.IsAtLeastVersion0_1_0

const (
	// ProjectServiceName is the fully-qualified name of the ProjectService service.
	ProjectServiceName = "runme.project.v1.ProjectService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// ProjectServiceLoadProcedure is the fully-qualified name of the ProjectService's Load RPC.
	ProjectServiceLoadProcedure = "/runme.project.v1.ProjectService/Load"
)

// ProjectServiceClient is a client for the runme.project.v1.ProjectService service.
type ProjectServiceClient interface {
	// Load creates a new project, walks it, and streams events
	// about found directories, files, and code blocks.
	Load(context.Context, *connect_go.Request[v1.LoadRequest]) (*connect_go.ServerStreamForClient[v1.LoadResponse], error)
}

// NewProjectServiceClient constructs a client for the runme.project.v1.ProjectService service. By
// default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses,
// and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewProjectServiceClient(httpClient connect_go.HTTPClient, baseURL string, opts ...connect_go.ClientOption) ProjectServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &projectServiceClient{
		load: connect_go.NewClient[v1.LoadRequest, v1.LoadResponse](
			httpClient,
			baseURL+ProjectServiceLoadProcedure,
			opts...,
		),
	}
}

// projectServiceClient implements ProjectServiceClient.
type projectServiceClient struct {
	load *connect_go.Client[v1.LoadRequest, v1.LoadResponse]
}

// Load calls runme.project.v1.ProjectService.Load.
func (c *projectServiceClient) Load(ctx context.Context, req *connect_go.Request[v1.LoadRequest]) (*connect_go.ServerStreamForClient[v1.LoadResponse], error) {
	return c.load.CallServerStream(ctx, req)
}

// ProjectServiceHandler is an implementation of the runme.project.v1.ProjectService service.
type ProjectServiceHandler interface {
	// Load creates a new project, walks it, and streams events
	// about found directories, files, and code blocks.
	Load(context.Context, *connect_go.Request[v1.LoadRequest], *connect_go.ServerStream[v1.LoadResponse]) error
}

// NewProjectServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewProjectServiceHandler(svc ProjectServiceHandler, opts ...connect_go.HandlerOption) (string, http.Handler) {
	projectServiceLoadHandler := connect_go.NewServerStreamHandler(
		ProjectServiceLoadProcedure,
		svc.Load,
		opts...,
	)
	return "/runme.project.v1.ProjectService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case ProjectServiceLoadProcedure:
			projectServiceLoadHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedProjectServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedProjectServiceHandler struct{}

func (UnimplementedProjectServiceHandler) Load(context.Context, *connect_go.Request[v1.LoadRequest], *connect_go.ServerStream[v1.LoadResponse]) error {
	return connect_go.NewError(connect_go.CodeUnimplemented, errors.New("runme.project.v1.ProjectService.Load is not implemented"))
}