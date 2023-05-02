package reports

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/data-drift/kpi-git-history/common"
	notion_database "github.com/data-drift/kpi-git-history/database/notion"
	"github.com/dstotijn/go-notion"
)

type httpTransport struct {
	w io.Writer
}

// RoundTrip implements http.RoundTripper. It multiplexes the read HTTP response
// data to an io.Writer for debugging.
func (t *httpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	res.Body = io.NopCloser(io.TeeReader(res.Body, t.w))

	return res, nil
}

func CreateReport(syncConfig common.SyncConfig, KPIInfo common.KPIInfo) error {

	reportNotionPageId, _ := notion_database.FindOrCreateReportPageId(syncConfig.NotionAPIKey, syncConfig.NotionDatabaseID, KPIInfo.KPIName)
	fmt.Println(reportNotionPageId)

	params := notion.CreatePageParams{
		Children: []notion.Block{
			notion.Heading1Block{
				RichText: []notion.RichText{
					{
						Text: &notion.Text{
							Content: "Problem",
						},
					},
				},
			},
			notion.ParagraphBlock{
				RichText: []notion.RichText{
					{
						Text: &notion.Text{
							Content: "Why has the ",
						},
					},
					{
						Text: &notion.Text{
							Content: KPIInfo.KPIName,
						},
						Annotations: &notion.Annotations{
							Code: true,
						},
					},
					{
						Text: &notion.Text{
							Content: " changed from " + strconv.Itoa(KPIInfo.FirstRoundedKPI) + " to " + strconv.Itoa(KPIInfo.LastRoundedKPI) + " ?",
						},
					},
				},
			},
			notion.Heading1Block{
				RichText: []notion.RichText{
					{
						Text: &notion.Text{
							Content: "Root Cause Analysis",
						},
					},
				},
			},
			notion.Heading2Block{
				RichText: []notion.RichText{
					{
						Text: &notion.Text{
							Content: "Overview",
						},
					},
				},
			},
			notion.ParagraphBlock{
				RichText: []notion.RichText{
					{
						Text: &notion.Text{
							Content: KPIInfo.KPIName,
						},
						Annotations: &notion.Annotations{
							Code: true,
						},
					},
					{
						Text: &notion.Text{
							Content: " initial value was: ",
						},
					},
					{
						Text: &notion.Text{
							Content: strconv.Itoa(KPIInfo.FirstRoundedKPI),
						},
						Annotations: &notion.Annotations{
							Bold: true,
						},
					},
				},
			},
			notion.ParagraphBlock{
				RichText: []notion.RichText{
					{
						Text: &notion.Text{
							Content: KPIInfo.KPIName,
						},
						Annotations: &notion.Annotations{
							Code: true,
						},
					},
					{
						Text: &notion.Text{
							Content: " current value is: ",
						},
					},
					{
						Text: &notion.Text{
							Content: strconv.Itoa(KPIInfo.LastRoundedKPI),
						},
						Annotations: &notion.Annotations{
							Bold: true,
						},
					},
				},
			},
			notion.Heading2Block{
				RichText: []notion.RichText{
					{
						Text: &notion.Text{
							Content: "Timeline",
						},
					},
				},
			},
			notion.EmbedBlock{
				URL: KPIInfo.GraphQLURL,
			},
			notion.Heading2Block{
				RichText: []notion.RichText{
					{
						Text: &notion.Text{
							Content: "Changelog",
						},
					},
				},
			},
		},
	}
	var children []notion.Block
	for _, event := range KPIInfo.Events {
		paragraph := notion.ParagraphBlock{
			RichText: []notion.RichText{
				{
					Text: &notion.Text{
						Content: "🗓 Date " + time.Unix(event.CommitTimestamp, 0).Format("2006-01-02"),
					},
					Annotations: &notion.Annotations{
						Underline: true,
					},
				},
			},
		}
		bulletListFirstItem := notion.BulletedListItemBlock{
			RichText: []notion.RichText{
				{
					Text: &notion.Text{
						Content: "Impact: ",
					},
				},
				{
					Text: &notion.Text{
						Content: strconv.Itoa(event.Diff),
					},
					Annotations: &notion.Annotations{
						Bold: true,
						Color: func() notion.Color {
							if event.Diff < 0 {
								return notion.ColorOrange
							}
							return notion.ColorBlue
						}(),
					},
				},
			},
		}
		bulletListSecondItem := notion.BulletedListItemBlock{
			RichText: []notion.RichText{
				{
					Text: &notion.Text{
						Content: "Explanations: ",
					},
				},
			},
		}
		children = append(children, paragraph, bulletListFirstItem, bulletListSecondItem)
	}
	params.Children = append(params.Children, children...)

	err := notion_database.UpdateReport(syncConfig.NotionAPIKey, reportNotionPageId, params.Children)
	if err != nil {
		return fmt.Errorf("failed to create page: %v", err)
	}

	return nil
}
