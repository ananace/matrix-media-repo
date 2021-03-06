package previewers

import (
	bytes2 "bytes"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/ryanuber/go-glob"
	"github.com/sirupsen/logrus"
	"github.com/turt2live/matrix-media-repo/common"
	"github.com/turt2live/matrix-media-repo/common/config"
	"github.com/turt2live/matrix-media-repo/controllers/preview_controller/preview_types"
	"github.com/turt2live/matrix-media-repo/metrics"
	"github.com/turt2live/matrix-media-repo/util"
)

func GenerateCalculatedPreview(urlPayload *preview_types.UrlPayload, log *logrus.Entry) (preview_types.PreviewResult, error) {
	bytes, filename, contentType, contentLength, err := downloadRawContent(urlPayload, config.Get().UrlPreviews.FilePreviewTypes, log)
	if err != nil {
		log.Error("Error downloading content: " + err.Error())

		// Make sure the unsupported error gets passed through
		if err == preview_types.ErrPreviewUnsupported {
			return preview_types.PreviewResult{}, preview_types.ErrPreviewUnsupported
		}

		// We'll consider it not found for the sake of processing
		return preview_types.PreviewResult{}, common.ErrMediaNotFound
	}

	stream := util.BufferToStream(bytes2.NewBuffer(bytes))
	img := &preview_types.PreviewImage{
		Data:                stream,
		ContentType:         contentType,
		Filename:            filename,
		ContentLength:       int64(len(bytes)),
		ContentLengthHeader: contentLength,
	}

	description := ""
	filename = urlPayload.ParsedUrl.String()
	if img != nil && img.Filename != "" {
		filename = img.Filename
	} else {
		description = urlPayload.ParsedUrl.String()
	}

	// Clear the description so we don't duplicate the URL
	if description == filename {
		description = ""
	}

	result := &preview_types.PreviewResult{
		Type:        "", // intentionally empty
		Url:         urlPayload.ParsedUrl.String(),
		Title:       summarize(filename, config.Get().UrlPreviews.NumTitleWords, config.Get().UrlPreviews.MaxTitleLength),
		Description: summarize(description, config.Get().UrlPreviews.NumWords, config.Get().UrlPreviews.MaxLength),
		SiteName:    "", // intentionally empty
	}

	if glob.Glob("image/*", img.ContentType) {
		result.Image = img
	}

	metrics.UrlPreviewsGenerated.With(prometheus.Labels{"type": "calculated"}).Inc()
	return *result, nil
}
