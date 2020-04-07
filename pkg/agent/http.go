package agent

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/horizon/pkg/edgeservices/logs"
	"github.com/hashicorp/horizon/pkg/wire"
)

type httpHandler struct {
	url string
}

func HTTPHandler(url string) ServiceHandler {
	if !strings.HasPrefix(url, "http://") {
		url = "http://" + url
	}

	return &httpHandler{url}
}

func (h *httpHandler) HandleRequest(ctx context.Context, L hclog.Logger, sctx ServiceContext) error {
	req := sctx.Request()

	L.Info("request started", "method", req.Method, "path", req.Path)

	hreq, err := http.NewRequestWithContext(ctx, req.Method, h.url+req.Path, sctx.BodyReader())
	if err != nil {
		return err
	}
	hreq.URL.RawQuery = req.Query
	hreq.URL.Fragment = req.Fragment
	if req.Auth != nil {
		hreq.URL.User = url.UserPassword(req.Auth.User, req.Auth.Password)
	}

	hresp, err := http.DefaultClient.Do(hreq)
	if err != nil {
		return err
	}

	defer hresp.Body.Close()

	var resp wire.Response
	resp.Code = int32(hresp.StatusCode)

	for k, v := range hresp.Header {
		resp.Headers = append(resp.Headers, &wire.Header{
			Name:  k,
			Value: v,
		})
	}

	err = sctx.WriteMarshal(1, &resp)
	if err != nil {
		return err
	}

	n, _ := io.Copy(sctx.BodyWriter(), hresp.Body)

	L.Info("request ended", "size", n)

	var lm logs.Message
	lm.Timestamp = logs.Now()
	lm.Mesg = "performed request"
	lm.Attrs = []*logs.Attribute{
		{
			Key:  "method",
			Sval: req.Method,
		},
		{
			Key:  "path",
			Sval: req.Path,
		},
		{
			Key:  "response-code",
			Ival: int64(hresp.StatusCode),
		},
		{
			Key:  "body-size",
			Ival: int64(n),
		},
	}

	return sctx.Log(&lm)
}