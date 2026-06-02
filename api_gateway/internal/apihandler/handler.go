package apihandler

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"api_gateway/internal/graph"
	"api_gateway/internal/resolver"
)

// New returns a fully configured HTTP router.
func New(r *resolver.Resolver) *http.ServeMux {
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: r}))

	mux := http.NewServeMux()
	mux.Handle("/graphql", srv)
	mux.Handle("/", playground.Handler("API Gateway Playground", "/graphql"))
	return mux
}
