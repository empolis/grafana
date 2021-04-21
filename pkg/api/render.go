package api

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/rendering"
	"github.com/grafana/grafana/pkg/util"
)

func Render(hs *HTTPServer, c *models.ReqContext, pdf bool, landscape bool) {
	queryReader, err := util.NewURLQueryReader(c.Req.URL)
	if err != nil {
		c.Handle(hs.Cfg, 400, "Render parameters error", err)
		return
	}

	queryParams := fmt.Sprintf("?%s", c.Req.URL.RawQuery)

	width, err := strconv.Atoi(queryReader.Get("width", "800"))
	if err != nil {
		c.Handle(hs.Cfg, 400, "Render parameters error", fmt.Errorf("cannot parse width as int: %s", err))
		return
	}

	height, err := strconv.Atoi(queryReader.Get("height", "400"))
	if err != nil {
		c.Handle(hs.Cfg, 400, "Render parameters error", fmt.Errorf("cannot parse height as int: %s", err))
		return
	}

	timeout, err := strconv.Atoi(queryReader.Get("timeout", "60"))
	if err != nil {
		c.Handle(hs.Cfg, 400, "Render parameters error", fmt.Errorf("cannot parse timeout as int: %s", err))
		return
	}

	scale, err := strconv.ParseFloat(queryReader.Get("scale", "1"), 64)
	if err != nil {
		c.Handle(hs.Cfg, 400, "Render parameters error", fmt.Errorf("cannot parse scale as float: %s", err))
		return
	}

	headers := http.Header{}
	acceptLanguageHeader := c.Req.Header.Values("Accept-Language")
	if len(acceptLanguageHeader) > 0 {
		headers["Accept-Language"] = acceptLanguageHeader
	}

	result, err := hs.RenderService.Render(c.Req.Context(), rendering.Opts{
		Width:             width,
		Height:            height,
		Timeout:           time.Duration(timeout) * time.Second,
		OrgId:             c.OrgId,
		UserId:            c.UserId,
		OrgRole:           c.OrgRole,
		Path:              c.Params("*") + queryParams,
		Timezone:          queryReader.Get("tz", ""),
		Encoding:          queryReader.Get("encoding", ""),
		ConcurrentLimit:   hs.Cfg.RendererConcurrentRequestLimit,
		DeviceScaleFactor: scale,
		Pdf:               pdf,
		Landscape:         landscape,
		Headers:           headers,
	})
	if err != nil {
		if errors.Is(err, rendering.ErrTimeout) {
			c.Handle(hs.Cfg, 500, err.Error(), err)
			return
		}
		if errors.Is(err, rendering.ErrPhantomJSNotInstalled) {
			if strings.HasPrefix(runtime.GOARCH, "arm") {
				c.Handle(hs.Cfg, 500, "Rendering failed - PhantomJS isn't included in arm build per default", err)
			} else {
				c.Handle(hs.Cfg, 500, "Rendering failed - PhantomJS isn't installed correctly", err)
			}
			return
		}

		c.Handle(hs.Cfg, 500, "Rendering failed.", err)
		return
	}

	if pdf {
		c.Resp.Header().Set("Content-Type", "application/pdf")
	} else {
		c.Resp.Header().Set("Content-Type", "image/png")
	}
	http.ServeFile(c.Resp, c.Req.Request, result.FilePath)
}

func (hs *HTTPServer) RenderToPng(c *models.ReqContext) {
	Render(hs, c, false, false)
}

func (hs *HTTPServer) RenderToPdf(c *models.ReqContext) {
	Render(hs, c, true, false)
}

func (hs *HTTPServer) RenderToPdfLandscape(c *models.ReqContext) {
	Render(hs, c, true, true)
}
