package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"waf-admin/internal/config"
	"waf-admin/internal/models"

	"github.com/sirupsen/logrus"
)

type LogsService struct {
	config *config.Config
	logger *logrus.Logger
}

func NewLogsService(cfg *config.Config, logger *logrus.Logger) *LogsService {
	return &LogsService{
		config: cfg,
		logger: logger,
	}
}

func (s *LogsService) SearchLogs(ctx context.Context, query models.LogQuery) (*models.LogSearchResult, error) {
	// Build VictoriaLogs query
	logSQL := s.buildLogSQL(query)
	
	u, err := url.Parse(s.config.Logs.VictoriaLogsURL + "/select/logsql/query")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("query", logSQL)
	q.Set("limit", fmt.Sprintf("%d", query.Limit))
	q.Set("offset", fmt.Sprintf("%d", query.Offset))
	q.Set("start", query.TimeRange.Start.Format(time.RFC3339))
	q.Set("end", query.TimeRange.End.Format(time.RFC3339))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("victoria logs returned status %d", resp.StatusCode)
	}

	var vlResult VictoriaLogsResult
	if err := json.NewDecoder(resp.Body).Decode(&vlResult); err != nil {
		return nil, err
	}

	result := &models.LogSearchResult{
		TimeRange: query.TimeRange,
		Total:     vlResult.Total,
		Entries:   make([]models.LogEntry, 0, len(vlResult.Logs)),
	}

	for _, log := range vlResult.Logs {
		entry := models.LogEntry{
			Timestamp: log.Timestamp,
			Message:   log.Message,
			Fields:    log.Fields,
		}

		// Extract common fields
		if host, ok := log.Fields["host"].(string); ok {
			entry.Host = host
		}
		if status, ok := log.Fields["status"].(float64); ok {
			entry.Status = int(status)
		}
		if ruleID, ok := log.Fields["rule_id"].(string); ok {
			entry.RuleID = ruleID
		}
		if clientIP, ok := log.Fields["remote_addr"].(string); ok {
			entry.ClientIP = clientIP
		}
		if path, ok := log.Fields["path"].(string); ok {
			entry.Path = path
		}
		if method, ok := log.Fields["method"].(string); ok {
			entry.Method = method
		}

		result.Entries = append(result.Entries, entry)
	}

	return result, nil
}

func (s *LogsService) buildLogSQL(query models.LogQuery) string {
	var conditions []string

	if query.Query != "" {
		conditions = append(conditions, query.Query)
	}

	// Add time-based conditions
	if query.TimeRange.Start.Unix() > 0 && query.TimeRange.End.Unix() > 0 {
		conditions = append(conditions, fmt.Sprintf("_time:%s..%s", 
			query.TimeRange.Start.Format(time.RFC3339),
			query.TimeRange.End.Format(time.RFC3339)))
	}

	if len(conditions) == 0 {
		return "*"
	}

	return strings.Join(conditions, " AND ")
}

func (s *LogsService) GetLogFilters() map[string][]string {
	return map[string][]string{
		"status": {"200", "403", "404", "500", "502", "503"},
		"method": {"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"},
		"rule_id": {"942100", "942110", "942120", "913100", "913110"}, // Common CRS rule IDs
	}
}

type VictoriaLogsResult struct {
	Total  int `json:"total"`
	Logs   []struct {
		Timestamp time.Time              `json:"_time"`
		Message   string                 `json:"_msg"`
		Fields    map[string]interface{} `json:"_fields"`
	} `json:"logs"`
}