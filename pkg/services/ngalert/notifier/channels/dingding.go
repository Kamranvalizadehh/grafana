package channels

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/alertmanager/types"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/notifications"
)

const defaultDingdingMsgType = "link"

// NewDingDingNotifier is the constructor for the Dingding notifier
func NewDingDingNotifier(model *NotificationChannelConfig, ns notifications.WebhookSender, t *template.Template) (*DingDingNotifier, error) {
	if model.Settings == nil {
		return nil, receiverInitError{Cfg: *model, Reason: "no settings supplied"}
	}
	if valid, err := ValidateContactPointReceiver(model.Type, model.Settings); err != nil || !valid {
		return nil, receiverInitError{Cfg: *model, Reason: err.Error()}
	}

	return &DingDingNotifier{
		Base: NewBase(&models.AlertNotification{
			Uid:                   model.UID,
			Name:                  model.Name,
			Type:                  model.Type,
			DisableResolveMessage: model.DisableResolveMessage,
			Settings:              model.Settings,
		}),
		MsgType: model.Settings.Get("msgType").MustString(defaultDingdingMsgType),
		URL:     model.Settings.Get("url").MustString(),
		Message: model.Settings.Get("message").MustString(`{{ template "default.message" .}}`),
		log:     log.New("alerting.notifier.dingding"),
		tmpl:    t,
		ns:      ns,
	}, nil
}

// DingDingNotifier is responsible for sending alert notifications to ding ding.
type DingDingNotifier struct {
	*Base
	MsgType string
	URL     string
	Message string
	tmpl    *template.Template
	ns      notifications.WebhookSender
	log     log.Logger
}

// Notify sends the alert notification to dingding.
func (dd *DingDingNotifier) Notify(ctx context.Context, as ...*types.Alert) (bool, error) {
	dd.log.Info("Sending dingding")

	ruleURL := joinUrlPath(dd.tmpl.ExternalURL.String(), "/alerting/list", dd.log)

	q := url.Values{
		"pc_slide": {"false"},
		"url":      {ruleURL},
	}

	// Use special link to auto open the message url outside of Dingding
	// Refer: https://open-doc.dingtalk.com/docs/doc.htm?treeId=385&articleId=104972&docType=1#s9
	messageURL := "dingtalk://dingtalkclient/page/link?" + q.Encode()

	var tmplErr error
	tmpl, _ := TmplText(ctx, dd.tmpl, as, dd.log, &tmplErr)

	message := tmpl(dd.Message)
	title := tmpl(DefaultMessageTitleEmbed)

	var bodyMsg map[string]interface{}
	if tmpl(dd.MsgType) == "actionCard" {
		bodyMsg = map[string]interface{}{
			"msgtype": "actionCard",
			"actionCard": map[string]string{
				"text":        message,
				"title":       title,
				"singleTitle": "More",
				"singleURL":   messageURL,
			},
		}
	} else {
		link := map[string]string{
			"text":       message,
			"title":      title,
			"messageUrl": messageURL,
		}

		bodyMsg = map[string]interface{}{
			"msgtype": "link",
			"link":    link,
		}
	}

	u := tmpl(dd.URL)
	if tmplErr != nil {
		dd.log.Warn("failed to template DingDing message", "err", tmplErr.Error())
	}

	body, err := json.Marshal(bodyMsg)
	if err != nil {
		return false, err
	}

	cmd := &models.SendWebhookSync{
		Url:  u,
		Body: string(body),
	}

	if err := dd.ns.SendWebhookSync(ctx, cmd); err != nil {
		return false, fmt.Errorf("send notification to dingding: %w", err)
	}

	return true, nil
}

func (dd *DingDingNotifier) SendResolved() bool {
	return !dd.GetDisableResolveMessage()
}
