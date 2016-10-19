package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph-go/pkg/lsp"

	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

func serveRepoDefLanding(w http.ResponseWriter, r *http.Request) error {
	repo, repoRev, err := handlerutil.GetRepoAndRev(r.Context(), mux.Vars(r))
	if err != nil {
		return errors.Wrap(err, "GetRepoAndRev")
	}

	// Parse query parameters.
	file := r.URL.Query().Get("file")
	line, err := strconv.Atoi(r.URL.Query().Get("line"))
	if err != nil {
		return errors.Wrap(err, "parsing line query param")
	}
	character, err := strconv.Atoi(r.URL.Query().Get("character"))
	if err != nil {
		return errors.Wrap(err, "parsing character query param")
	}

	// TODO: figure out how to handle other languages here.
	language := "go"

	// Lookup the symbol's information by performing textDocument/definition
	// and then looking through workspace/symbol results for the definition.
	rootPath := "git://" + repo.URI + "?" + repoRev.CommitID
	var locations []lsp.Location
	err = lspClientRequest(r.Context(), language, rootPath, "textDocument/definition", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: rootPath + "#" + file},
		Position:     lsp.Position{Line: line, Character: character},
	}, &locations)
	if len(locations) == 0 {
		return fmt.Errorf("textDocument/definition returned zero locations")
	}
	uri, err := url.Parse(locations[0].URI)
	if err != nil {
		return errors.Wrap(err, "parsing definition URL")
	}

	// Query workspace symbols.
	withoutFile := *uri
	withoutFile.Fragment = ""
	var symbols []lsp.SymbolInformation
	err = lspClientRequest(r.Context(), language, withoutFile.String(), "workspace/symbol", lsp.WorkspaceSymbolParams{
		// TODO(slimsag): before merge, performance for golang/go here is not
		// good. Allow specifying file URIs as a query filter. Sucks a bit that
		// textDocument/definition won't give us the Name/ContainerName that we
		// need!
		Query: "", // all symbols
	}, &symbols)

	// Find the matching symbol.
	var symbol *lsp.SymbolInformation
	for _, sym := range symbols {
		if sym.Location.URI != locations[0].URI {
			continue
		}
		if sym.Location.Range.Start.Line != locations[0].Range.Start.Line {
			continue
		}
		if sym.Location.Range.Start.Character != locations[0].Range.Start.Character {
			continue
		}
		symbol = &sym
		break
	}
	if symbol == nil {
		return fmt.Errorf("could not finding matching symbol info")
	}

	legacyURL, err := legacyDefLandingURL(*symbol)
	if err != nil {
		return errors.Wrap(err, "legacyDefLandingURL")
	}

	w.Header().Set("cache-control", "private, max-age=60")
	return writeJSON(w, &struct {
		URL string
	}{
		URL: legacyURL,
	})
}

// legacyDefLandingURL creates a relative URL to the legacy def landing page
// route. For example:
//
//  /github.com/gorilla/mux/-/info/GoPackage/NewRouter/mux/-/github.com/gorilla/mux
//  /github.com/golang/go/-/info/GoPackage/Encode/Encoder/-/encoding/gob
//
func legacyDefLandingURL(s lsp.SymbolInformation) (string, error) {
	uri, err := url.Parse(s.Location.URI)
	if err != nil {
		return "", err
	}

	defPath := s.Name
	if s.ContainerName != "" {
		defPath = s.ContainerName + "/" + s.Name
	}

	repo := uri.Host + uri.Path
	unit := uri.Host + path.Join(uri.Path, path.Dir(uri.Fragment))
	if repo == "github.com/golang/go" {
		// Special case golang/go to emit just "encoding/json" for the path "github.com/golang/go/src/encoding/json"
		unit = strings.TrimPrefix(path.Dir(uri.Fragment), "src/")
	}

	return router.Rel.URLToDefLanding(graph.DefKey{
		Repo:     repo,
		CommitID: "",
		UnitType: "GoPackage",
		Unit:     unit,
		Path:     defPath,
	}).String(), nil
}

// lspClientRequest performs a one-shot LSP client request for the specified
// method (e.g. "textDocument/definition") and stores the results in the given
// pointer value.
func lspClientRequest(ctx context.Context, mode, rootPath, method string, params, results interface{}) error {
	// Connect to the xlang proxy.
	c, err := xlang.NewDefaultClient()
	if err != nil {
		return errors.Wrap(err, "new xlang client")
	}
	defer c.Close()

	// Initialize the connection.
	err = c.Call(ctx, "initialize", xlang.ClientProxyInitializeParams{
		InitializeParams: lsp.InitializeParams{
			RootPath: rootPath,
		},
		Mode: mode,
	}, nil)
	if err != nil {
		return errors.Wrap(err, "lsp initialize")
	}

	// Perform the request.
	err = c.Call(ctx, method, params, results)
	if err != nil {
		return errors.Wrap(err, "lsp textDocument/definition")
	}

	// Shutdown the connection.
	err = c.Call(ctx, "shutdown", nil, nil)
	if err != nil {
		return errors.Wrap(err, "shutdown")
	}
	err = c.Notify(ctx, "exit", nil)
	if err != nil {
		return errors.Wrap(err, "exit")
	}
	return nil
}
