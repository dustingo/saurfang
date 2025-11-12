package pkg

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"text/template"
	"time"

	"saurfang/internal/config"
	"saurfang/internal/models/task"

	"github.com/hashicorp/consul/api"
	nomadapi "github.com/hashicorp/nomad/api"
)

// CustomTaskService handles custom task execution logic
type CustomTaskService struct {
	NomadClient *nomadapi.Client
}

// NewCustomTaskService creates a new CustomTaskService instance
func NewCustomTaskService() (*CustomTaskService, error) {
	return &CustomTaskService{
		NomadClient: config.NomadCli,
	}, nil
}

// ExecuteCustomTaskAsync executes a custom task asynchronously
func (s *CustomTaskService) ExecuteCustomTaskAsync(customTask *task.CustomTask, execution *task.CustomTaskExecution) {
	defer func() {
		execution.EndTime = &time.Time{}
		*execution.EndTime = time.Now()
		config.DB.Save(execution)
	}()

	// Parse target hosts
	targetHosts := s.parseTargetHosts(customTask.TargetHosts)
	if len(targetHosts) == 0 {
		execution.Status = "failed"
		execution.ErrorMsg = "no target hosts specified"
		return
	}

	// Generate separate Job configuration for each target host
	var dispatchResults []string
	for _, host := range targetHosts {
		// Create task copy for each host, update target_hosts parameter
		hostTask := *customTask
		hostTask.TargetHosts = host // Single host name

		// Generate Nomad Job configuration
		jobSpec, err := s.generateNomadJobSpec(&hostTask)
		if err != nil {
			slog.Error("Failed to generate job spec for host", "host", host, "error", err)
			dispatchResults = append(dispatchResults, fmt.Sprintf("Failed to generate job spec for %s: %v", host, err))
			continue
		}

		// Debug: print generated Job configuration
		slog.Info("Generated Nomad Job spec for host", "host", host, "job_spec", jobSpec)

		// Save Job HCL to Consul (using host-specific key)
		consulKey := fmt.Sprintf("custom_job/%d_%s", customTask.ID, host)
		if err = s.saveJobHCLToConsulWithKey(jobSpec, consulKey); err != nil {
			slog.Warn("Failed to save job HCL to Consul", "host", host, "error", err)
		}

		// Parse and register Nomad Job
		job, err := s.NomadClient.Jobs().ParseHCL(jobSpec, true)
		if err != nil {
			slog.Error("Failed to parse job spec for host", "host", host, "error", err)
			dispatchResults = append(dispatchResults, fmt.Sprintf("Failed to parse job spec for %s: %v", host, err))
			continue
		}

		// Set Job ID (include host information)
		jobID := fmt.Sprintf("custom-task-%d-%d-%s", customTask.ID, execution.ID, host)
		job.ID = &jobID

		// Register Job
		resp, _, err := s.NomadClient.Jobs().Register(job, &nomadapi.WriteOptions{})
		if err != nil {
			slog.Error("Failed to register job for host", "host", host, "error", err)
			dispatchResults = append(dispatchResults, fmt.Sprintf("Failed to register job for %s: %v", host, err))
		} else {
			dispatchResults = append(dispatchResults, fmt.Sprintf("Registered job for %s: %s (eval: %s)", host, jobID, resp.EvalID))
		}
	}

	execution.NomadJobID = fmt.Sprintf("custom-task-%d-%d", customTask.ID, execution.ID)
	execution.Status = "success"
	execution.Result = fmt.Sprintf("Jobs registered for all hosts. Results: %s", strings.Join(dispatchResults, "; "))

	// Update task last execution time
	now := time.Now()
	customTask.LastRun = &now
	config.DB.Save(&customTask)
}

// generateNomadJobSpec generates Nomad Job configuration
func (s *CustomTaskService) generateNomadJobSpec(customTask *task.CustomTask) (string, error) {
	// Parse parameters
	var params map[string]interface{}
	if customTask.Parameters != "" {
		if err := json.Unmarshal([]byte(customTask.Parameters), &params); err != nil {
			return "", fmt.Errorf("failed to parse parameters: %v", err)
		}
	}

	// Render script template
	renderedScript, err := s.renderScriptTemplate(customTask.Script, params)
	if err != nil {
		return "", fmt.Errorf("failed to render script template: %v", err)
	}

	// Generate different Job configuration based on script type
	switch customTask.ScriptType {
	case "bash", "shell":
		return s.generateShellJob(customTask, renderedScript)
	case "python":
		return s.generatePythonJob(customTask, renderedScript)
	default:
		return "", fmt.Errorf("unsupported script type: %s", customTask.ScriptType)
	}
}

// renderScriptTemplate renders script template with parameters
func (s *CustomTaskService) renderScriptTemplate(scriptTemplate string, params map[string]interface{}) (string, error) {
	tmpl, err := template.New("script").Parse(scriptTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse script template: %v", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, params); err != nil {
		return "", fmt.Errorf("failed to execute script template: %v", err)
	}

	return buf.String(), nil
}

// generateShellJob generates shell job configuration
func (s *CustomTaskService) generateShellJob(customTask *task.CustomTask, script string) (string, error) {
	// Parse target hosts for node constraints
	targetHosts := s.parseTargetHosts(customTask.TargetHosts)
	nodeConstraints := s.generateNodeConstraints(targetHosts)

	// Format script for HCL
	formattedScript, err := s.formatHCL(script)
	if err != nil {
		return "", fmt.Errorf("failed to format script: %v", err)
	}

	jobSpec := fmt.Sprintf(`
job "custom-task-%d" {
  datacenters = ["dc1"]
  type = "batch"
  
  %s
  
  group "custom-task-group" {
    count = 1
    
    task "custom-task" {
      driver = "raw_exec"
      
      config {
        command = "/bin/bash"
        args = ["-c", %q]
      }
      
      resources {
        cpu    = 500
        memory = 512
      }
      
      %s
    }
  }
}`, customTask.ID, nodeConstraints, formattedScript, s.generateTimeoutConfig(customTask.Timeout))

	return jobSpec, nil
}

// generatePythonJob generates Python job configuration
func (s *CustomTaskService) generatePythonJob(customTask *task.CustomTask, script string) (string, error) {
	// Parse target hosts for node constraints
	targetHosts := s.parseTargetHosts(customTask.TargetHosts)
	nodeConstraints := s.generateNodeConstraints(targetHosts)

	// Format script for HCL
	formattedScript, err := s.formatHCL(script)
	if err != nil {
		return "", fmt.Errorf("failed to format script: %v", err)
	}

	jobSpec := fmt.Sprintf(`
job "custom-task-%d" {
  datacenters = ["dc1"]
  type = "batch"
  
  %s
  
  group "custom-task-group" {
    count = 1
    
    task "custom-task" {
      driver = "raw_exec"
      
      config {
        command = "/usr/bin/python3"
        args = ["-c", %q]
      }
      
      resources {
        cpu    = 500
        memory = 512
      }
      
      %s
    }
  }
}`, customTask.ID, nodeConstraints, formattedScript, s.generateTimeoutConfig(customTask.Timeout))

	return jobSpec, nil
}

// parseTargetHosts parses target hosts string into slice
func (s *CustomTaskService) parseTargetHosts(targetHostsStr string) []string {
	if targetHostsStr == "" {
		return []string{}
	}

	// Split by comma and clean up whitespace
	hosts := strings.Split(targetHostsStr, ",")
	var cleanHosts []string
	for _, host := range hosts {
		cleanHost := strings.TrimSpace(host)
		if cleanHost != "" {
			cleanHosts = append(cleanHosts, cleanHost)
		}
	}

	return cleanHosts
}

// generateNodeConstraints generates node constraints for Nomad job
func (s *CustomTaskService) generateNodeConstraints(targetHosts []string) string {
	if len(targetHosts) == 0 {
		return ""
	}

	// Create constraint for specific nodes
	// var constraints []string
	// for _, host := range targetHosts {
	// 	constraints = append(constraints, fmt.Sprintf(`"${attr.unique.hostname}" == "%s"`, host))
	// }

	return fmt.Sprintf(`
  constraint {
    attribute = "${attr.unique.hostname}"
    operator  = "regexp"
    value     = "(%s)"
  }`, strings.Join(targetHosts, "|"))
}

// generateTimeoutConfig generates timeout configuration
func (s *CustomTaskService) generateTimeoutConfig(timeout int) string {
	if timeout <= 0 {
		return ""
	}

	return fmt.Sprintf(`
      kill_timeout = "%ds"`, timeout)
}

// formatHCL formats content for HCL
func (s *CustomTaskService) formatHCL(hclContent string) (string, error) {
	// Escape quotes and newlines for HCL string
	formatted := strings.ReplaceAll(hclContent, `"`, `\"`)
	formatted = strings.ReplaceAll(formatted, "\n", "\\n")
	return formatted, nil
}

// saveJobHCLToConsulWithKey saves job HCL to Consul with specific key
func (s *CustomTaskService) saveJobHCLToConsulWithKey(jobSpec string, key string) error {
	_, err := config.ConsulCli.KV().Put(&api.KVPair{
		Key:   key,
		Value: []byte(jobSpec),
	}, nil)
	return err
}
