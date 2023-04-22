package handlers

import (
	"context"
	"testing"

	"github.com/size12/url-shortener/internal/config"
	"github.com/size12/url-shortener/internal/storage"
	pb "github.com/size12/url-shortener/pkg/grpc"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestGRPCServer(t *testing.T) {
	cfg := config.GetTestConfig()
	s, _ := storage.NewMapStorage(cfg)

	handlers := NewService(cfg, s)
	// get new shortener server
	server := NewShortenerServer(cfg, handlers)

	assert.Equal(t, &ShortenerServer{
		cfg:     cfg,
		service: handlers,
	}, server)

}

func TestGRPCHandlers(t *testing.T) {
	cfg := config.GetTestConfig()
	s, err := storage.NewMapStorage(cfg)
	assert.NoError(t, err)
	handlers := NewService(cfg, s)
	server := NewShortenerServer(cfg, handlers)

	in := &pb.Link{
		LongUrl: "https://yandex.ru",
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{"userID": "cookieUser12"}))

	// create short.
	out, err := server.CreateShort(ctx, in)
	assert.NoError(t, err)
	assert.Equal(t, &pb.Link{
		ShortUrl: cfg.BaseURL + "/" + "1",
		Id:       "1",
	}, out)

	assert.Contains(t, s.Locations, out.Id)

	copyMapLoc := s.Locations

	// create bad short.
	in.LongUrl = "rkgrekgeg"
	_, err = server.CreateShort(ctx, in)
	assert.Error(t, err)
	assert.Equal(t, copyMapLoc, s.Locations)

	// create short which is already in storage.
	in.LongUrl = "https://yandex.ru"
	out, err = server.CreateShort(ctx, in)
	assert.NoError(t, err)
	assert.Equal(t, &pb.Link{
		ShortUrl: cfg.BaseURL + "/" + "1",
		Id:       "1",
	}, out)

	assert.Contains(t, s.Locations, out.Id)
	assert.Equal(t, copyMapLoc, s.Locations)

	// check ping.
	_, err = server.Ping(ctx, &emptypb.Empty{})
	assert.NoError(t, err)

	// delete short.
	_, err = server.Delete(ctx, out)
	assert.NoError(t, err)
	assert.Equal(t, true, s.Deleted["1"])

	// delete again short.
	_, err = server.Delete(ctx, out)
	assert.NoError(t, err)
	assert.Equal(t, true, s.Deleted["1"])

	// delete non exists short.
	out.Id = "3"
	_, err = server.Delete(ctx, out)
	assert.NoError(t, err)

	// get statistic.
	stat, err := server.GetStatistics(ctx, &emptypb.Empty{})
	assert.NoError(t, err)
	assert.Equal(t, &pb.Statistic{Urls: 1, Users: 1}, stat)

	// get non existed long.
	_, err = server.GetLong(ctx, out)
	assert.Error(t, err)
	// get existed long.
	out.Id = "1"
	long, err := server.GetLong(ctx, out)
	assert.Error(t, err)
	assert.Equal(t, &pb.Link{
		LongUrl: "https://yandex.ru",
	}, long)

	// batch short.
	links := &pb.Batch{}
	links.Result = []*pb.Link{
		{
			CorrelationId: "1",
			LongUrl:       "https://google.com",
		},
		{

			CorrelationId: "2",
			LongUrl:       "https://youtube.com",
		},
	}

	result, err := server.BatchShort(ctx, links)
	assert.NoError(t, err)

	assert.Equal(t,
		&pb.Batch{
			Result: []*pb.Link{
				{
					CorrelationId: "1",
					ShortUrl:      cfg.BaseURL + "/2",
				},
				{

					CorrelationId: "2",
					ShortUrl:      cfg.BaseURL + "/3",
				},
			},
		}, result)

	history, err := server.GetHistory(ctx, &emptypb.Empty{})
	assert.NoError(t, err)

	assert.Equal(t, &pb.Batch{
		Result: []*pb.Link{
			{

				LongUrl:  "https://yandex.ru",
				ShortUrl: cfg.BaseURL + "/1",
			},
			{
				LongUrl:  "https://google.com",
				ShortUrl: cfg.BaseURL + "/2",
			},
			{

				LongUrl:  "https://youtube.com",
				ShortUrl: cfg.BaseURL + "/3",
			},
		},
	}, history)

}
