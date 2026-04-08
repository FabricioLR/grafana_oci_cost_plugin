package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
	"net/http"

	"github.com/patrickmn/go-cache"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/transoft/oci-cost/pkg/models"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/usageapi"

	//"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

// Make sure Datasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler interfaces. Plugin should not implement all these
// interfaces - only those which are required for a particular task.
var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

// NewDatasource creates a new datasource instance.
func NewDatasource(_ context.Context, _ backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &Datasource{}, nil
}

// Datasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type Datasource struct{}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *Datasource) Dispose() {
	// Clean up datasource instance resources.
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).

type queryModel struct {
    Detailing  string `json:"det"`
	Type string `json:"type"`
}

var c = cache.New(1*time.Hour, 10*time.Minute)
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 120*time.Second)
    defer cancel()
	plugin_config, err := models.LoadPluginSettings(*req.PluginContext.DataSourceInstanceSettings)
	response := backend.NewQueryDataResponse()

	if err != nil {
		return response, nil
	}

	if err != nil {
		return response, nil
	}

	if plugin_config.Secrets.UserOCID == "" {
		return response, nil
	}

	if plugin_config.Secrets.TenancyOCID == "" {
		return response, nil
	}

	if plugin_config.Secrets.Fingerprint == "" {
		return response, nil
	}

	if plugin_config.Secrets.Region == "" {
		return response, nil
	}

	if plugin_config.Secrets.PrivateKey == "" {
		return response, nil
	}

	config := common.NewRawConfigurationProvider(
		plugin_config.Secrets.TenancyOCID,
		plugin_config.Secrets.UserOCID,
		plugin_config.Secrets.Region,
		plugin_config.Secrets.Fingerprint,
		strings.ReplaceAll(plugin_config.Secrets.PrivateKey, "\\n", "\n"),
		nil,
	)

	client, err := usageapi.NewUsageapiClientWithConfigurationProvider(config)
    if err != nil {
        return response, nil
    }

	httpClient := &http.Client{
		Timeout: 120 * time.Second, // Timeout total da requisição HTTP
	}
	client.HTTPClient = httpClient

	firstQuery := req.Queries[0]
	var model queryModel

	err = json.Unmarshal(firstQuery.JSON, &model)
	if err != nil {
		response.Responses[firstQuery.RefID] = backend.DataResponse{Error: err}
		return response, nil
	}

	from := firstQuery.TimeRange.From
	to := firstQuery.TimeRange.To
	//startTime := common.SDKTime{Time: from.Truncate(24 * time.Hour)}
	//endTime := common.SDKTime{Time: to.Truncate(24 * time.Hour)}
	startTime := common.SDKTime{
		Time: time.Date(from.Year(), from.Month(), from.Day() + 1, 0, 0, 0, 0, time.UTC),
	}
	endTime := common.SDKTime{
		Time: time.Date(to.Year(), to.Month(), to.Day() + 1, 0, 0, 0, 0, time.UTC),
	}

	//if to.After(startTime.Time) {
	//	endTime.Time = endTime.Time.AddDate(0, 0, 2) 
	//}

	var res usageapi.RequestSummarizedUsagesResponse
	const cacheTimeFormat = "2006-01-02T15"
	cacheKey := fmt.Sprintf("usage_%s_%s_%s", startTime.Format(cacheTimeFormat), endTime.Format(cacheTimeFormat), model.Detailing)
	if cachedData, found := c.Get(cacheKey); found {
		//log.DefaultLogger.Info("------------------------------------------------cache foi usado------------------------------------------------")
		res = cachedData.(usageapi.RequestSummarizedUsagesResponse)
	} else {
		if (model.Detailing == "Dias"){
			request := usageapi.RequestSummarizedUsagesRequest{
				RequestSummarizedUsagesDetails: usageapi.RequestSummarizedUsagesDetails{
					TenantId:         common.String(plugin_config.Secrets.TenancyOCID),
					TimeUsageStarted: &startTime,
					TimeUsageEnded:   &endTime,
					Granularity:      usageapi.RequestSummarizedUsagesDetailsGranularityDaily,
					GroupBy:          []string{"service", "tag"},
				},
			}

			res, err = client.RequestSummarizedUsages(ctxWithTimeout, request)
			if err != nil {
				response.Responses[firstQuery.RefID] = backend.DataResponse{Error: err}
				return response, nil
			}
		} else if (model.Detailing == "Meses") {
			request := usageapi.RequestSummarizedUsagesRequest{
				RequestSummarizedUsagesDetails: usageapi.RequestSummarizedUsagesDetails{
					TenantId:         common.String(plugin_config.Secrets.TenancyOCID),
					TimeUsageStarted: &startTime,
					TimeUsageEnded:   &endTime,
					Granularity:      usageapi.RequestSummarizedUsagesDetailsGranularityMonthly,
					GroupBy:          []string{"service", "tag"},
				},
			}

			res, err = client.RequestSummarizedUsages(ctxWithTimeout, request)
			if err != nil {
				response.Responses[firstQuery.RefID] = backend.DataResponse{Error: err}
				return response, nil
			}
		}

		c.Set(cacheKey, res, cache.DefaultExpiration)
	}

    for _, q := range req.Queries {
		pivotData := make(map[time.Time]map[string]float64)
		instancesFound := make(map[string]bool)

		for _, item := range res.Items {
			if item.Tags == nil || len(item.Tags) == 0 {
				continue 
			}

			encontrouTag := false
			valorEncontrado := "null"
			if (model.Type == "database"){
				if (item.Service != nil && *item.Service != "Database"){
					continue
				}
				for _, t := range item.Tags {
					namespace := ""
					if t.Namespace != nil { namespace = *t.Namespace }
					
					key := ""
					if t.Key != nil { key = *t.Key }

					//transoft id_transoft server1-2-3-4...
					if namespace == "Transnet" && key == "banco" {
						encontrouTag = true
						if t.Value != nil {
							valorEncontrado = *t.Value
						}
						break
					}
				}
			} else if (model.Type == "compute"){
				if (item.Service != nil && *item.Service != "Compute"){
					continue
				}
				for _, t := range item.Tags {
					namespace := ""
					if t.Namespace != nil { namespace = *t.Namespace }
					
					key := ""
					if t.Key != nil { key = *t.Key }

					//transoft id_transoft server1-2-3-4...
					if namespace == "transoft" && key == "id_transoft" {
						if t.Value != nil {
							if strings.HasPrefix(*t.Value, "server") {
								encontrouTag = true
								valorEncontrado = *t.Value
								break
							}
						}
					}
				}
			} else if (model.Type == "compute-all"){
				if (item.Service != nil && *item.Service != "Compute"){
					continue
				}
				for _, t := range item.Tags {
					namespace := ""
					if t.Namespace != nil { namespace = *t.Namespace }
					
					key := ""
					if t.Key != nil { key = *t.Key }

					//transoft id_transoft server1-2-3-4...
					if namespace == "transoft" && key == "id_transoft" {
						if t.Value != nil {
							if !strings.HasPrefix(*t.Value, "server") {
								encontrouTag = true
								valorEncontrado = *t.Value
								break
							}
						}
					}
				}
			}

			if !encontrouTag {
				continue
			}

			t := item.TimeUsageStarted.Time

			amount := 0.0
			if item.ComputedAmount != nil {
				amount = float64(*item.ComputedAmount)
			}

			if _, ok := pivotData[t]; !ok {
				pivotData[t] = make(map[string]float64)
			}
			pivotData[t][valorEncontrado] += amount
			instancesFound[valorEncontrado] = true
		}

		var instanceCols []string
		frame := data.NewFrame("database_costs")
		frame.Fields = append(frame.Fields, data.NewField("Time", nil, []time.Time{}))

		for id := range instancesFound {
			instanceCols = append(instanceCols, id)
			frame.Fields = append(frame.Fields, data.NewField(id, nil, []float64{}))
		}

		var dates []time.Time
		for d := range pivotData {
			dates = append(dates, d)
		}
		sort.Slice(dates, func(i, j int) bool { return dates[i].Before(dates[j]) })

		for _, date := range dates {
			row := make([]interface{}, len(instanceCols)+1)
			row[0] = date
			for i, id := range instanceCols {
				row[i+1] = pivotData[date][id]
			}
			frame.AppendRow(row...)
		}

        response.Responses[q.RefID] = backend.DataResponse{Frames: data.Frames{frame}}
    }

    return response, nil
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *Datasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	res := &backend.CheckHealthResult{}
	config, err := models.LoadPluginSettings(*req.PluginContext.DataSourceInstanceSettings)

	if err != nil {
		res.Status = backend.HealthStatusError
		res.Message = "Unable to load settings"
		return res, nil
	}

	if config.Secrets.UserOCID == "" {
		res.Status = backend.HealthStatusError
		res.Message = "User Ocid is missing"
		return res, nil
	}

	if config.Secrets.TenancyOCID == "" {
		res.Status = backend.HealthStatusError
		res.Message = "Tenancy Ocid is missing"
		return res, nil
	}

	if config.Secrets.Fingerprint == "" {
		res.Status = backend.HealthStatusError
		res.Message = "Fingerprint is missing"
		return res, nil
	}

	if config.Secrets.Region == "" {
		res.Status = backend.HealthStatusError
		res.Message = "Region is missing"
		return res, nil
	}

	if config.Secrets.PrivateKey == "" {
		res.Status = backend.HealthStatusError
		res.Message = "Private Key is missing"
		return res, nil
	}

	ociconfig := common.NewRawConfigurationProvider(
		config.Secrets.TenancyOCID,
		config.Secrets.UserOCID,
		config.Secrets.Region,
		config.Secrets.Fingerprint,
		strings.ReplaceAll(config.Secrets.PrivateKey, "\\n", "\n"),
		nil,
	)

	_, err = usageapi.NewUsageapiClientWithConfigurationProvider(ociconfig)
    if err != nil {
        res.Status = backend.HealthStatusError
		res.Message = "Cannot create OCI client connection"
		return res, nil
    }

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Data source is working",
	}, nil
}
