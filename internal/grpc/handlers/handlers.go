package handlers

import (
	"context"
	"errors"
	"net/url"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/ivanmyagkov/shortener.git/internal/grpc/proto"
	"github.com/ivanmyagkov/shortener.git/internal/interfaces"
	"github.com/ivanmyagkov/shortener.git/internal/utils"
)

type GRPCHandler struct {
	storage  interfaces.Storage
	cfg      interfaces.Config
	user     interfaces.Users
	inWorker interfaces.InWorker
}

//	New is function to set server settings.
func NewGRPCHandler(storage interfaces.Storage, config interfaces.Config, user interfaces.Users, inWorker interfaces.InWorker) *GRPCHandler {
	return &GRPCHandler{
		storage:  storage,
		cfg:      config,
		user:     user,
		inWorker: inWorker,
	}
}

// GetPingDB handles PSQL DB pinging to check connection status.
func (h *GRPCHandler) GetPingDB() (*pb.PingResponse, error) {
	err := h.storage.Ping()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var response pb.PingResponse
	return &response, nil
}

// GetStats is handler to get stats
func (h *GRPCHandler) GetStats() (*pb.GetStatsResponse, error) {
	stat, err := h.storage.GetStats()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	response := pb.GetStatsResponse{
		Users: int64(stat.Users),
		Urls:  int64(stat.URLs),
	}
	return &response, nil
}

// GetURL is handler to get stats
func (h *GRPCHandler) GetURL(request *pb.GetURLRequest) (*pb.GetURLResponse, error) {
	shortURL := request.ShortUrlId
	URL, err := h.storage.GetURL(shortURL)
	if err != nil {
		if errors.Is(err, interfaces.ErrWasDeleted) {
			return nil, status.Error(codes.NotFound, err.Error())
		} else if errors.Is(err, interfaces.ErrNotFound) {
			return nil, status.Error(codes.Internal, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	response := pb.GetURLResponse{
		RedirectTo: URL,
	}
	return &response, nil
}

// PostURL method to crerate short URL
func (h *GRPCHandler) PostURL(ctx context.Context, request *pb.PostURLRequest) (*pb.PostURLResponse, error) {
	token := ctx.Value(interfaces.UserIDCtxName)
	userID, _ := h.user.ReadSessionID(token.(string))
	shortURL, err := h.shortenURL(userID, request.BaseUrl)
	if err != nil {
		if errors.Is(err, interfaces.ErrAlreadyExists) {
			shortURL = utils.NewURL(h.cfg.HostName(), shortURL)
			response := pb.PostURLResponse{
				ShortUrl: shortURL,
			}
			return &response, status.Error(codes.AlreadyExists, `Entry already exists and was returned in response body`)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	response := pb.PostURLResponse{
		ShortUrl: shortURL,
	}
	return &response, nil
}

// GetURLsByUserID method to get ULS by user
func (h *GRPCHandler) GetURLsByUserID(ctx context.Context) (*pb.GetURLsByUserIDResponse, error) {
	token := ctx.Value(interfaces.UserIDCtxName)
	userID, _ := h.user.ReadSessionID(token.(string))
	URLs, err := h.storage.GetAllURLsByUserID(userID)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	response := pb.GetURLsByUserIDResponse{}
	for _, v := range URLs {
		responseURL := pb.ResponseURLs{
			BaseUrl:  v.BaseURL,
			ShortUrl: v.ShortURL,
		}
		response.ResponseUrls = append(response.ResponseUrls, &responseURL)
	}
	return &response, nil
}

// PostURLBatch method to create array shorten URLs
func (h *GRPCHandler) PostURLBatch(ctx context.Context, request *pb.PostURLBatchRequest) (*pb.PostURLBatchResponse, error) {
	token := ctx.Value(interfaces.UserIDCtxName)
	userID, _ := h.user.ReadSessionID(token.(string))
	response := pb.PostURLBatchResponse{}
	for _, requestBatchURL := range request.RequestUrls {
		var responseBatchURL pb.PostURLBatch
		responseBatchURL.CorrelationId = requestBatchURL.CorrelationId
		var err error
		responseBatchURL.Url, err = h.shortenURL(userID, requestBatchURL.Url)
		if err != nil {
			if errors.Is(err, interfaces.ErrAlreadyExists) {
				return nil, status.Error(codes.AlreadyExists, err.Error())
			}
			return nil, status.Error(codes.Internal, err.Error())
		}
		response.ResponseUrls = append(response.ResponseUrls, &responseBatchURL)
	}
	return &response, nil
}

// DeleteURLBatch method to delete array shorten URLs
func (h *GRPCHandler) DeleteURLBatch(ctx context.Context, request *pb.DeleteURLBatchRequest) (*pb.DeleteURLBatchResponse, error) {
	token := ctx.Value(interfaces.UserIDCtxName)
	userID, _ := h.user.ReadSessionID(token.(string))
	var model interfaces.Task
	model.ID = userID
	for _, deleteURL := range request.RequestUrls.Urls {
		model.ShortURL = deleteURL
		h.inWorker.Do(model)
	}
	var response pb.DeleteURLBatchResponse
	return &response, nil
}

// shortenURL - Auxiliary link shortening function
func (h GRPCHandler) shortenURL(userID, URL string) (string, error) {
	_, err := url.ParseRequestURI(URL)
	if err != nil {
		return "", err
	}
	shortURL := utils.MD5([]byte(URL))
	err = h.storage.SetShortURL(userID, shortURL, URL)
	if err != nil {
		if errors.Is(err, interfaces.ErrAlreadyExists) {
			shortURL = utils.NewURL(h.cfg.HostName(), shortURL)
			return shortURL, interfaces.ErrAlreadyExists
		} else {
			return "", err
		}
	}
	shortURL = utils.NewURL(h.cfg.HostName(), shortURL)
	return shortURL, nil
}
