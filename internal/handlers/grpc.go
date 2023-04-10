package handlers

import (
	"context"

	"github.com/size12/url-shortener/internal/config"
	"github.com/size12/url-shortener/internal/storage"
	pb "github.com/size12/url-shortener/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ShortenerServer struct {
	pb.UnimplementedShortenerServer
	cfg     config.Config
	service *Service
}

func NewShortenerServer(cfg config.Config, service *Service) *ShortenerServer {
	return &ShortenerServer{
		cfg:     cfg,
		service: service,
	}
}

func (server *ShortenerServer) Ping(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	empty := &emptypb.Empty{}
	err := server.service.CheckPing()
	if err != nil {
		return empty, status.Error(codes.Unavailable, "Storage doesn't response.")
	}
	return empty, nil
}

func (server *ShortenerServer) CreateShort(ctx context.Context, in *pb.Link) (*pb.Link, error) {
	result := &pb.Link{}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "wrong metadata")
	}

	if len(md.Get("userID")) == 0 {
		return nil, status.Error(codes.Unknown, "wrong metadata")
	}

	id, err := server.service.ShortSingleURL(md.Get("userID")[0], in.LongUrl)

	if err != storage.Err409 && err != nil {
		return result, err
	}

	result.ShortUrl = server.cfg.BaseURL + "/" + id
	result.Id = id

	if err == storage.Err409 {
		err = nil
	}

	return result, err
}

func (server *ShortenerServer) GetStatistics(ctx context.Context, _ *emptypb.Empty) (*pb.Statistic, error) {
	result := &pb.Statistic{}
	stat, err := server.service.GetStatistic()
	result.Users = uint32(stat.Users)
	result.Urls = uint32(stat.Urls)

	return result, err
}

func (server *ShortenerServer) GetLong(ctx context.Context, in *pb.Link) (*pb.Link, error) {
	result := &pb.Link{}
	long, err := server.service.GetLongURL(in.Id)
	if err == storage.Err404 {
		return nil, status.Error(codes.NotFound, "Link not in storage")
	}
	result.LongUrl = long
	return result, err
}

func (server *ShortenerServer) Delete(ctx context.Context, in *pb.Link) (*emptypb.Empty, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "wrong metadata")
	}

	if len(md.Get("userID")) == 0 {
		return nil, status.Error(codes.Unknown, "wrong metadata")
	}

	userID := md.Get("userID")[0]

	err := server.service.DeleteURL(userID, []string{in.Id})
	return nil, err
}

func (server *ShortenerServer) GetHistory(ctx context.Context, in *emptypb.Empty) (*pb.Batch, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md.Get("userID")) == 0 {
		return nil, status.Error(codes.Unknown, "wrong metadata")
	}

	userID := md.Get("userID")[0]

	history, err := server.service.GetHistory(userID)

	if err != nil {
		return nil, err
	}

	result := &pb.Batch{}
	for _, elem := range history {
		result.Result = append(result.Result, &pb.Link{
			LongUrl:  elem.LongURL,
			ShortUrl: elem.ShortURL,
		})
	}

	return result, nil
}

func (server *ShortenerServer) BatchShort(ctx context.Context, in *pb.Batch) (*pb.Batch, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md.Get("userID")) == 0 {
		return nil, status.Error(codes.Unknown, "wrong metadata")
	}

	userID := md.Get("userID")[0]

	result := &pb.Batch{}
	query := make([]storage.BatchJSON, 0, len(in.Result))

	for _, url := range in.Result {
		query = append(query, storage.BatchJSON{
			CorrelationID: url.CorrelationId,
			URL:           url.LongUrl,
		})
	}

	urls, err := server.service.ShortURLs(userID, query)

	if err == storage.Err409 {
		err = nil
	}

	if err != nil {
		return nil, err
	}

	result.Result = make([]*pb.Link, 0, len(urls))

	for _, url := range urls {
		result.Result = append(result.Result, &pb.Link{
			CorrelationId: url.CorrelationID,
			ShortUrl:      url.ShortURL,
		})
	}

	return result, nil
}
