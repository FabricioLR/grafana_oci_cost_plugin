package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
	"net/url"
	"regexp"

	"github.com/patrickmn/go-cache"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/transoft/oci-cost/pkg/models"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/identity"
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
func NewDatasource(_ context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	config, err := models.LoadPluginSettings(settings)
	if err != nil {
		return nil, err
	}

	return &Datasource{
		config: config,
	}, nil
}

// Datasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type Datasource struct {
	config *models.PluginSettings
}

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
	Detailing string `json:"det"`
	Type      string `json:"type"`
	Service   string `json:"service"`
	Namespace string `json:"namespace"`
	Tag       string `json:"tag_key"`
	Value     string `json:"tag_value"`
}

var c = cache.New(1*time.Hour, 10*time.Minute)

func matches(pattern string, value string) bool {
	if strings.HasPrefix(pattern, "!") {
        regexReal := strings.TrimPrefix(pattern, "!")
        matched, err := regexp.MatchString(regexReal, value)
		if err == nil && !matched{
			return true
		}

        return pattern == value
    } else {
		matched, err := regexp.MatchString(pattern, value)
		if err == nil && matched {
			return true
		}

		return pattern == value
	}
}

func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	response := backend.NewQueryDataResponse()

	//log.DefaultLogger.Info("%+v", d.config)

	if d.config.Secrets.UserOCID == "" || d.config.Secrets.TenancyOCID == "" || d.config.Secrets.Fingerprint == "" || d.config.Secrets.Region == "" || d.config.Secrets.PrivateKey == "" {
		return response, nil
	}

	config := common.NewRawConfigurationProvider(
		d.config.Secrets.TenancyOCID,
		d.config.Secrets.UserOCID,
		d.config.Secrets.Region,
		d.config.Secrets.Fingerprint,
		strings.ReplaceAll(d.config.Secrets.PrivateKey, "\\n", "\n"),
		nil,
	)

	client, err := usageapi.NewUsageapiClientWithConfigurationProvider(config)
	if err != nil {
		return response, nil
	}

	httpClient := &http.Client{
		Timeout: 120 * time.Second,
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
	startTime := common.SDKTime{
		Time: time.Date(from.Year(), from.Month(), from.Day()+1, 0, 0, 0, 0, time.UTC),
	}
	endTime := common.SDKTime{
		Time: time.Date(to.Year(), to.Month(), to.Day()+1, 0, 0, 0, 0, time.UTC),
	}

	var res usageapi.RequestSummarizedUsagesResponse
	const cacheTimeFormat = "2006-01-02T15"
	cacheKey := fmt.Sprintf("usage_%s_%s_%s", startTime.Format(cacheTimeFormat), endTime.Format(cacheTimeFormat), model.Detailing)
	if cachedData, found := c.Get(cacheKey); found {
		res = cachedData.(usageapi.RequestSummarizedUsagesResponse)
	} else {
		if model.Detailing == "Dias" {
			request := usageapi.RequestSummarizedUsagesRequest{
				RequestSummarizedUsagesDetails: usageapi.RequestSummarizedUsagesDetails{
					TenantId:         common.String(d.config.Secrets.TenancyOCID),
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
		} else if model.Detailing == "Meses" {
			request := usageapi.RequestSummarizedUsagesRequest{
				RequestSummarizedUsagesDetails: usageapi.RequestSummarizedUsagesDetails{
					TenantId:         common.String(d.config.Secrets.TenancyOCID),
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
			encontrouTag := false
			valorEncontrado := "null"
			if model.Service != "All" {
				if item.Service != nil && *item.Service != model.Service {
					continue
				}

				if item.Tags == nil || len(item.Tags) == 0 {
					continue
				}

				for _, t := range item.Tags {
					namespace := ""
					if t.Namespace != nil {
						namespace = *t.Namespace
					}

					key := ""
					if t.Key != nil {
						key = *t.Key
					}

					if (model.Namespace != "All" && matches(model.Namespace, namespace)) {
						if (model.Tag != "All" && matches(model.Tag, key)){
							if (model.Value != "All" && t.Value != nil && matches(model.Value, *t.Value)){
								encontrouTag = true
								valorEncontrado = *t.Value
								break
							} else if (model.Value == "All" && t.Value != nil){
								encontrouTag = true
								valorEncontrado = *t.Value
								break
							}
						} else if (model.Tag == "All"){
							if (model.Value != "All" && t.Value != nil && matches(model.Value, *t.Value)){
								encontrouTag = true
								valorEncontrado = *t.Value
								break
							} else if (model.Value == "All" && t.Value != nil){
								encontrouTag = true
								valorEncontrado = *t.Value
								break
							}
						}
					} else if (model.Namespace == "All"){
						if (model.Tag != "All" && matches(model.Tag, key)){
							if (model.Value != "All" && t.Value != nil && matches(model.Value, *t.Value)){
								encontrouTag = true
								valorEncontrado = *t.Value
								break
							} else if (model.Value == "All" && t.Value != nil){
								encontrouTag = true
								valorEncontrado = *t.Value
								break
							}
						} else if (model.Tag == "All"){
							if (model.Value != "All" && t.Value != nil && matches(model.Value, *t.Value)){
								encontrouTag = true
								valorEncontrado = *t.Value
								break
							} else if (model.Value == "All" && t.Value != nil){
								encontrouTag = true
								valorEncontrado = *t.Value
								break
							}
						}
					}
				}
			} else if model.Service == "All" {
				if (model.Namespace != "All" || model.Tag != "All" || model.Value != "All"){
					if item.Tags == nil || len(item.Tags) == 0 {
						continue
					}

					for _, t := range item.Tags {
						namespace := ""
						if t.Namespace != nil {
							namespace = *t.Namespace
						}

						key := ""
						if t.Key != nil {
							key = *t.Key
						}

						if (model.Namespace != "All" && matches(model.Namespace, namespace)) {
							if (model.Tag != "All" && matches(model.Tag, key)){
								if (model.Value != "All" && t.Value != nil && matches(model.Value, *t.Value)){
									encontrouTag = true
									valorEncontrado = *t.Value
									break
								} else if (model.Value == "All" && t.Value != nil){
									encontrouTag = true
									valorEncontrado = *t.Value
									break
								}
							} else if (model.Tag == "All"){
								if (model.Value != "All" && t.Value != nil && matches(model.Value, *t.Value)){
									encontrouTag = true
									valorEncontrado = *t.Value
									break
								} else if (model.Value == "All" && t.Value != nil){
									encontrouTag = true
									valorEncontrado = *t.Value
									break
								}
							}
						} else if (model.Namespace == "All"){
							if (model.Tag != "All" && matches(model.Tag, key)){
								if (model.Value != "All" && t.Value != nil && matches(model.Value, *t.Value)){
									encontrouTag = true
									valorEncontrado = *t.Value
									break
								} else if (model.Value == "All" && t.Value != nil){
									encontrouTag = true
									valorEncontrado = *t.Value
									break
								}
							} else if (model.Tag == "All"){
								if (model.Value != "All" && t.Value != nil && matches(model.Value, *t.Value)){
									encontrouTag = true
									valorEncontrado = *t.Value
									break
								} else if (model.Value == "All" && t.Value != nil){
									encontrouTag = true
									valorEncontrado = *t.Value
									break
								}
							}
						}
					}
				} else {
					encontrouTag = true
					if item.Service != nil {
						valorEncontrado = *item.Service
					} else {
						valorEncontrado = "Other"
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

	if config.Secrets.UserOCID == "" || config.Secrets.TenancyOCID == "" || config.Secrets.Fingerprint == "" || config.Secrets.Region == "" || config.Secrets.PrivateKey == "" {
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

func writeResponse(sender backend.CallResourceResponseSender, data interface{}) error {
	jsonRes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return sender.Send(&backend.CallResourceResponse{
		Status: http.StatusOK,
		Body:   jsonRes,
	})
}

func (d *Datasource) fetchTagNamespaces(ctx context.Context) ([]string, error) {
	if d.config.Secrets.UserOCID == "" || d.config.Secrets.TenancyOCID == "" || d.config.Secrets.Fingerprint == "" || d.config.Secrets.Region == "" || d.config.Secrets.PrivateKey == "" {
		return nil, nil
	}

	config := common.NewRawConfigurationProvider(
		d.config.Secrets.TenancyOCID,
		d.config.Secrets.UserOCID,
		d.config.Secrets.Region,
		d.config.Secrets.Fingerprint,
		strings.ReplaceAll(d.config.Secrets.PrivateKey, "\\n", "\n"),
		nil,
	)

	request := identity.ListTagNamespacesRequest{
		CompartmentId: common.String(d.config.Secrets.TenancyOCID),
		IncludeSubcompartments: common.Bool(true),
		Limit: common.Int(100),
	}

	client, err := identity.NewIdentityClientWithConfigurationProvider(config)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar usage client: %w", err)
	}

	res, err := client.ListTagNamespaces(ctx, request)
	if err != nil {
		return nil, err
	}

	var namespaces []string
	for _, ns := range res.Items {
		namespaces = append(namespaces, *ns.Name)
	}
	namespaces = append(namespaces, "All")
	return namespaces, nil
}

func (d *Datasource) fetchTagsByNamespace(ctx context.Context, namespaceName string) ([]string, error) {
	if namespaceName == "" {
        return []string{}, nil
    }
	
	if d.config.Secrets.UserOCID == "" || d.config.Secrets.TenancyOCID == "" || d.config.Secrets.Fingerprint == "" || d.config.Secrets.Region == "" || d.config.Secrets.PrivateKey == "" {
		return nil, nil
	}

	config := common.NewRawConfigurationProvider(
		d.config.Secrets.TenancyOCID,
		d.config.Secrets.UserOCID,
		d.config.Secrets.Region,
		d.config.Secrets.Fingerprint,
		strings.ReplaceAll(d.config.Secrets.PrivateKey, "\\n", "\n"),
		nil,
	)

	var tags []string
	if (namespaceName == "All"){
		request1 := identity.ListTagNamespacesRequest{
			CompartmentId: common.String(d.config.Secrets.TenancyOCID),
			IncludeSubcompartments: common.Bool(true),
			Limit: common.Int(100),
		}

		client, err := identity.NewIdentityClientWithConfigurationProvider(config)
		if err != nil {
			return nil, fmt.Errorf("falha ao criar usage client: %w", err)
		}

		res, err := client.ListTagNamespaces(ctx, request1)
		if err != nil {
			return nil, err
		}

		for _, ns := range res.Items {
			request2 := identity.ListTagsRequest{
				TagNamespaceId: common.String(*ns.Name),
				Limit:          common.Int(100),
			}

			client, err := identity.NewIdentityClientWithConfigurationProvider(config)
			if err != nil {
				return nil, fmt.Errorf("falha ao criar usage client: %w", err)
			}

			res, err := client.ListTags(ctx, request2)
			if err != nil {
				return nil, err
			}

			for _, tag := range res.Items {
				if tag.Name != nil {
					tags = append(tags, *tag.Name)
				}
			}
		}
	} else {
		request := identity.ListTagsRequest{
			TagNamespaceId: common.String(namespaceName),
			Limit:          common.Int(100),
		}

		client, err := identity.NewIdentityClientWithConfigurationProvider(config)
		if err != nil {
			return nil, fmt.Errorf("falha ao criar usage client: %w", err)
		}

		res, err := client.ListTags(ctx, request)
		if err != nil {
			return nil, err
		}

		for _, tag := range res.Items {
			if tag.Name != nil {
				tags = append(tags, *tag.Name)
			}
		}
	}
	
	tags = append(tags, "All")
    return tags, nil
}

func (d *Datasource) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	var data []string
	var err error

	u, err := url.Parse(req.URL)
    if err != nil {
        return sender.Send(&backend.CallResourceResponse{Status: http.StatusBadRequest})
    }

	switch u.Path {
	case "namespaces":
		data, err = d.fetchTagNamespaces(ctx)
	case "tags":
		ns := u.Query().Get("namespace")
		data, err = d.fetchTagsByNamespace(ctx, ns)
	default:
		return sender.Send(&backend.CallResourceResponse{Status: http.StatusNotFound})
	}

	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte(err.Error()),
		})
	}

	return writeResponse(sender, data)
}
