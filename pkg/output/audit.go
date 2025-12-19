package output

import (
	"fmt"
	"os"
	"time"

	"github.com/aboigues/k8t/pkg/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// AuditLogger provides structured logging for cluster access audit trail (SR-004)
type AuditLogger struct {
	logger  *zap.Logger
	entries []types.AuditEntry
}

// NewAuditLogger creates a new audit logger
// Logs are written to stderr in simple parseable format (clarification Q5)
func NewAuditLogger(verbose bool) (*AuditLogger, error) {
	// Configure encoder for simple stdout/stderr parsing
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "",
		MessageKey:     "msg",
		StacktraceKey:  "",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   nil,
	}

	// Determine log level
	level := zapcore.InfoLevel
	if verbose {
		level = zapcore.DebugLevel
	}

	// Create core that writes to stderr
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(os.Stderr),
		level,
	)

	logger := zap.New(core)

	return &AuditLogger{
		logger:  logger,
		entries: make([]types.AuditEntry, 0),
	}, nil
}

// LogResourceAccess logs cluster resource access
func (a *AuditLogger) LogResourceAccess(resourceType, resourceName, namespace, operation string) {
	entry := types.AuditEntry{
		Timestamp:    a.now(),
		ResourceType: resourceType,
		ResourceName: resourceName,
		Namespace:    namespace,
		Operation:    operation,
	}

	a.entries = append(a.entries, entry)

	// Log to stderr
	a.logger.Info("cluster_access",
		zap.String("resource_type", resourceType),
		zap.String("resource_name", resourceName),
		zap.String("namespace", namespace),
		zap.String("operation", operation),
	)
}

// LogPodGet logs pod retrieval
func (a *AuditLogger) LogPodGet(podName, namespace string) {
	a.LogResourceAccess("pods", podName, namespace, "get")
}

// LogPodList logs pod listing
func (a *AuditLogger) LogPodList(namespace string) {
	a.LogResourceAccess("pods", "", namespace, "list")
}

// LogEventList logs event listing
func (a *AuditLogger) LogEventList(namespace string) {
	a.LogResourceAccess("events", "", namespace, "list")
}

// LogSecretGet logs secret retrieval (for imagePullSecrets validation)
func (a *AuditLogger) LogSecretGet(secretName, namespace string) {
	a.LogResourceAccess("secrets", secretName, namespace, "get")
}

// LogAnalysisStart logs the beginning of analysis
func (a *AuditLogger) LogAnalysisStart(targetType types.TargetType, targetName, namespace string) {
	a.logger.Info("analysis_start",
		zap.String("target_type", string(targetType)),
		zap.String("target_name", targetName),
		zap.String("namespace", namespace),
	)
}

// LogAnalysisComplete logs the completion of analysis
func (a *AuditLogger) LogAnalysisComplete(targetType types.TargetType, targetName, namespace string, findingsCount int) {
	a.logger.Info("analysis_complete",
		zap.String("target_type", string(targetType)),
		zap.String("target_name", targetName),
		zap.String("namespace", namespace),
		zap.Int("findings_count", findingsCount),
	)
}

// LogError logs an error message
func (a *AuditLogger) LogError(message string, err error) {
	a.logger.Error(message, zap.Error(err))
}

// LogWarning logs a warning message
func (a *AuditLogger) LogWarning(message string) {
	a.logger.Warn(message)
}

// LogDebug logs a debug message (only if verbose mode enabled)
func (a *AuditLogger) LogDebug(message string, fields ...zap.Field) {
	a.logger.Debug(message, fields...)
}

// GetAuditEntries returns all recorded audit entries
func (a *AuditLogger) GetAuditEntries() []types.AuditEntry {
	return a.entries
}

// Close flushes and closes the logger
func (a *AuditLogger) Close() error {
	return a.logger.Sync()
}

// now returns the current time (extracted for testing)
func (a *AuditLogger) now() time.Time {
	return time.Now()
}

// Simple utility for non-verbose logging
func SimpleLog(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}
