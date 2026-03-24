package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/mcpfleet/registry/internal/db"
)

type Handler struct {
	store *db.Store
}

func RegisterRoutes(api huma.API, store *db.Store) {
	h := &Handler{store: store}

	// Servers
	huma.Register(api, huma.Operation{
		OperationID: "list-servers",
		Method:      http.MethodGet,
		Path:        "/v1/servers",
		Summary:     "List MCP servers",
		Tags:        []string{"servers"},
	}, h.ListServers)

	huma.Register(api, huma.Operation{
		OperationID: "get-server",
		Method:      http.MethodGet,
		Path:        "/v1/servers/{id}",
		Summary:     "Get MCP server",
		Tags:        []string{"servers"},
	}, h.GetServer)

	huma.Register(api, huma.Operation{
		OperationID: "create-server",
		Method:      http.MethodPost,
		Path:        "/v1/servers",
		Summary:     "Create MCP server",
		Tags:        []string{"servers"},
		DefaultStatus: http.StatusCreated,
	}, h.CreateServer)

	huma.Register(api, huma.Operation{
		OperationID: "update-server",
		Method:      http.MethodPut,
		Path:        "/v1/servers/{id}",
		Summary:     "Update MCP server",
		Tags:        []string{"servers"},
	}, h.UpdateServer)

	huma.Register(api, huma.Operation{
		OperationID: "delete-server",
		Method:      http.MethodDelete,
		Path:        "/v1/servers/{id}",
		Summary:     "Delete MCP server",
		Tags:        []string{"servers"},
		DefaultStatus: http.StatusNoContent,
	}, h.DeleteServer)

	// Tokens
	huma.Register(api, huma.Operation{
		OperationID: "list-tokens",
		Method:      http.MethodGet,
		Path:        "/v1/tokens",
		Summary:     "List auth tokens",
		Tags:        []string{"tokens"},
	}, h.ListTokens)

	huma.Register(api, huma.Operation{
		OperationID: "create-token",
		Method:      http.MethodPost,
		Path:        "/v1/tokens",
		Summary:     "Create auth token",
		Tags:        []string{"tokens"},
		DefaultStatus: http.StatusCreated,
	}, h.CreateToken)

	huma.Register(api, huma.Operation{
		OperationID: "delete-token",
		Method:      http.MethodDelete,
		Path:        "/v1/tokens/{id}",
		Summary:     "Delete auth token",
		Tags:        []string{"tokens"},
		DefaultStatus: http.StatusNoContent,
	}, h.DeleteToken)
}

// ---- Input/Output types ----

type ServerOutput struct {
	Body *db.Server
}

type ServersOutput struct {
	Body []db.Server
}

type ServerInput struct {
	ID   string `path:"id"`
	Body struct {
		Name        string            `json:"name" minLength:"1"`
		Description string            `json:"description"`
		Command     string            `json:"command" minLength:"1"`
		Args        []string          `json:"args"`
		Env         map[string]string `json:"env"`
		Tags        []string          `json:"tags"`
	}
}

type IDInput struct {
	ID string `path:"id"`
}

type TokenOutput struct {
	Body *db.CreateTokenResult
}

type TokensOutput struct {
	Body []db.Token
}

type CreateTokenInput struct {
	Body struct {
		Name string `json:"name" minLength:"1"`
	}
}

// ---- Handlers ----

func (h *Handler) ListServers(ctx context.Context, _ *struct{}) (*ServersOutput, error) {
	servers, err := h.store.ListServers(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to list servers", err)
	}
	if servers == nil {
		servers = []db.Server{}
	}
	return &ServersOutput{Body: servers}, nil
}

func (h *Handler) GetServer(ctx context.Context, input *IDInput) (*ServerOutput, error) {
	srv, err := h.store.GetServer(ctx, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to get server", err)
	}
	if srv == nil {
		return nil, huma.Error404NotFound("server not found")
	}
	return &ServerOutput{Body: srv}, nil
}

func (h *Handler) CreateServer(ctx context.Context, input *struct {
	Body struct {
		Name        string            `json:"name" minLength:"1"`
		Description string            `json:"description"`
		Command     string            `json:"command" minLength:"1"`
		Args        []string          `json:"args"`
		Env         map[string]string `json:"env"`
		Tags        []string          `json:"tags"`
	}
}) (*ServerOutput, error) {
	srv := &db.Server{
		Name:        input.Body.Name,
		Description: input.Body.Description,
		Command:     input.Body.Command,
		Args:        input.Body.Args,
		Env:         input.Body.Env,
		Tags:        input.Body.Tags,
	}
	if err := h.store.CreateServer(ctx, srv); err != nil {
		return nil, huma.Error500InternalServerError("failed to create server", err)
	}
	return &ServerOutput{Body: srv}, nil
}

func (h *Handler) UpdateServer(ctx context.Context, input *ServerInput) (*ServerOutput, error) {
	existing, err := h.store.GetServer(ctx, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to get server", err)
	}
	if existing == nil {
		return nil, huma.Error404NotFound("server not found")
	}
	existing.Name = input.Body.Name
	existing.Description = input.Body.Description
	existing.Command = input.Body.Command
	existing.Args = input.Body.Args
	existing.Env = input.Body.Env
	existing.Tags = input.Body.Tags
	if err := h.store.UpdateServer(ctx, existing); err != nil {
		return nil, huma.Error500InternalServerError("failed to update server", err)
	}
	return &ServerOutput{Body: existing}, nil
}

func (h *Handler) DeleteServer(ctx context.Context, input *IDInput) (*struct{}, error) {
	if err := h.store.DeleteServer(ctx, input.ID); err != nil {
		return nil, huma.Error500InternalServerError("failed to delete server", err)
	}
	return nil, nil
}

func (h *Handler) ListTokens(ctx context.Context, _ *struct{}) (*TokensOutput, error) {
	tokens, err := h.store.ListTokens(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to list tokens", err)
	}
	if tokens == nil {
		tokens = []db.Token{}
	}
	return &TokensOutput{Body: tokens}, nil
}

func (h *Handler) CreateToken(ctx context.Context, input *CreateTokenInput) (*TokenOutput, error) {
	result, err := h.store.CreateToken(ctx, input.Body.Name)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to create token", err)
	}
	return &TokenOutput{Body: result}, nil
}

func (h *Handler) DeleteToken(ctx context.Context, input *IDInput) (*struct{}, error) {
	if err := h.store.DeleteToken(ctx, input.ID); err != nil {
		return nil, huma.Error500InternalServerError("failed to delete token", err)
	}
	return nil, nil
}
