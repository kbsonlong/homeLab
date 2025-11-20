package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"waf-admin/internal/config"
	"waf-admin/internal/models"

	"github.com/sirupsen/logrus"
)

type MetricsService struct {
	config *config.Config
	logger *logrus.Logger
}

func NewMetricsService(cfg *config.Config, logger *logrus.Logger) *MetricsService {
	return &MetricsService{
		config: cfg,
		logger: logger,
	}
}

func (s *MetricsService) GetMetricsSummary(ctx context.Context, timeRange models.TimeRange) (*models.MetricsSummary, error) {
	summary := &models.MetricsSummary{
		TimeRange: timeRange,
	}

	// Query total requests
	totalQuery := fmt.Sprintf(`sum(rate(nginx_ingress_controller_requests[%s]))`, s.formatDuration(timeRange.End.Sub(timeRange.Start)))
	totalResult, err := s.queryVictoriaMetrics(ctx, totalQuery, timeRange)
	if err != nil {
		s.logger.Warnf("Failed to query total requests: %v", err)
	} else if len(totalResult.Data.Result) > 0 {
		summary.TotalRequests = int64(totalResult.Data.Result[0].Value[1].(float64))
	}

	// Query 4xx requests
	fourXXQuery := fmt.Sprintf(`sum(rate(nginx_ingress_controller_requests{status=~"4.."}[%s]))`, s.formatDuration(timeRange.End.Sub(timeRange.Start)))
	fourXXResult, err := s.queryVictoriaMetrics(ctx, fourXXQuery, timeRange)
	if err != nil {
		s.logger.Warnf("Failed to query 4xx requests: %v", err)
	} else if len(fourXXResult.Data.Result) > 0 {
		summary.Status4xx = int64(fourXXResult.Data.Result[0].Value[1].(float64))
	}

	// Query 5xx requests
	fiveXXQuery := fmt.Sprintf(`sum(rate(nginx_ingress_controller_requests{status=~"5.."}[%s]))`, s.formatDuration(timeRange.End.Sub(timeRange.Start)))
	fiveXXResult, err := s.queryVictoriaMetrics(ctx, fiveXXQuery, timeRange)
	if err != nil {
		s.logger.Warnf("Failed to query 5xx requests: %v", err)
	} else if len(fiveXXResult.Data.Result) > 0 {
		summary.Status5xx = int64(fiveXXResult.Data.Result[0].Value[1].(float64))
	}

	// Query 403 requests (WAF blocked)
	fourZeroThreeQuery := fmt.Sprintf(`sum(rate(nginx_ingress_controller_requests{status="403"}[%s]))`, s.formatDuration(timeRange.End.Sub(timeRange.Start)))
	fourZeroThreeResult, err := s.queryVictoriaMetrics(ctx, fourZeroThreeQuery, timeRange)
	if err != nil {
		s.logger.Warnf("Failed to query 403 requests: %v", err)
	} else if len(fourZeroThreeResult.Data.Result) > 0 {
		summary.Status403 = int64(fourZeroThreeResult.Data.Result[0].Value[1].(float64))
		summary.WAFBlocked = summary.Status403
	}

	// Query top hosts
	topHostsQuery := fmt.Sprintf(`sum(rate(nginx_ingress_controller_requests[%s])) by (host)`, s.formatDuration(timeRange.End.Sub(timeRange.Start)))
	topHostsResult, err := s.queryVictoriaMetrics(ctx, topHostsQuery, timeRange)
	if err != nil {
		s.logger.Warnf("Failed to query top hosts: %v", err)
	} else {
		for _, result := range topHostsResult.Data.Result {
			if host, ok := result.Metric["host"]; ok {
				requests := int64(result.Value[1].(float64))
				hostMetric := models.HostMetrics{
					Host:     host,
					Requests: requests,
				}

				// Query blocked requests for this host
				blockedQuery := fmt.Sprintf(`sum(rate(nginx_ingress_controller_requests{host="%s",status="403"}[%s]))`, host, s.formatDuration(timeRange.End.Sub(timeRange.Start)))
				blockedResult, err := s.queryVictoriaMetrics(ctx, blockedQuery, timeRange)
				if err == nil && len(blockedResult.Data.Result) > 0 {
					hostMetric.Blocked = int64(blockedResult.Data.Result[0].Value[1].(float64))
				}

				if requests > 0 {
					hostMetric.ErrorRate = float64(hostMetric.Blocked) / float64(requests) * 100
				}

				summary.TopHosts = append(summary.TopHosts, hostMetric)
			}
		}
	}

	return summary, nil
}

func (s *MetricsService) queryVictoriaMetrics(ctx context.Context, query string, timeRange models.TimeRange) (*VMQueryResult, error) {
	u, err := url.Parse(s.config.Metrics.VictoriaMetricsURL + "/api/v1/query")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("query", query)
	q.Set("time", timeRange.End.Format(time.RFC3339))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("victoria metrics returned status %d", resp.StatusCode)
	}

	var result VMQueryResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *MetricsService) formatDuration(d time.Duration) string {
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh", int(d.Hours()))
}

type VMQueryResult struct {
	Status string `json:"status"`
	Data   struct {
		Result []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		} `json:"result"`
	} `json:"data"`
}