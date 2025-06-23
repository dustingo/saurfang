package tools

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/gofiber/fiber/v3"
	"html/template"
	"io"
	"os"
	"saurfang/internal/config"

	"github.com/apenella/go-ansible/v2/pkg/execute"
	"github.com/apenella/go-ansible/v2/pkg/execute/measure"
	"github.com/apenella/go-ansible/v2/pkg/playbook"
	"saurfang/internal/models/datasource"
	"saurfang/internal/models/serverconfig"
	"saurfang/internal/models/task.go"
	"strings"
	"time"
)

type FlushingWriter struct {
	ctx fiber.Ctx
}

func (fw *FlushingWriter) Write(p []byte) (n int, err error) {
	n, err = fw.ctx.Write(p)
	if bw, ok := fw.ctx.Response().BodyWriter().(*bufio.Writer); ok {
		bw.Flush()
	}
	return n, err
}

// RunAnsibleDeployPlaybooks 发布进程
func RunAnsibleDeployPlaybooks(hosts string, ds *datasource.SaurfangDatasources, pts *task.SaurfangPublishtasks, cnf map[string]serverconfig.Configs, ctx fiber.Ctx) error {
	flushingWriter := &FlushingWriter{ctx: ctx}
	plbs, err := newCommandPlaybook(ds, pts, cnf)
	if err != nil {
		return err
	}
	ansiblePlaybookOptions := &playbook.AnsiblePlaybookOptions{
		Connection: "ssh",
		Inventory:  fmt.Sprintf("%s%s", hosts, ","),
	}
	playbookTasks := playbook.NewAnsiblePlaybookCmd(
		playbook.WithPlaybooks(plbs),
		playbook.WithPlaybookOptions(ansiblePlaybookOptions),
	)
	exec := measure.NewExecutorTimeMeasurement(
		execute.NewDefaultExecute(
			execute.WithCmd(playbookTasks),
			execute.WithWrite(flushingWriter),
		),
	)
	err = exec.Execute(context.TODO())
	if err != nil {
		return err
	}
	return nil
}

// RunAnsibleOpsPlaybooks 执行普通任务
func RunAnsibleOpsPlaybooks(hosts, keys string, ctx fiber.Ctx) error {
	flushingWriter := &FlushingWriter{ctx: ctx}
	plbs, err := newOpsTaskPlaybook(keys)
	if err != nil {
		return err
	}
	ansiblePlaybookOptions := &playbook.AnsiblePlaybookOptions{
		Connection: "ssh",
		Inventory:  fmt.Sprintf("%s%s", hosts, ","),
	}
	playbookTasks := playbook.NewAnsiblePlaybookCmd(
		playbook.WithPlaybooks(plbs...),
		playbook.WithPlaybookOptions(ansiblePlaybookOptions),
	)
	exec := measure.NewExecutorTimeMeasurement(
		execute.NewDefaultExecute(
			execute.WithCmd(playbookTasks),
			execute.WithWrite(flushingWriter),
		),
	)
	err = exec.Execute(context.TODO())
	if err != nil {
		return err
	}
	return nil
}
func RunAnsibleOpsPlaybooksByCron(hosts, keys string) error {
	plbs, err := newOpsTaskPlaybook(keys)
	if err != nil {
		return err
	}
	ansiblePlaybookOptions := &playbook.AnsiblePlaybookOptions{
		Connection: "ssh",
		Inventory:  fmt.Sprintf("%s%s", hosts, ","),
	}
	playbookTasks := playbook.NewAnsiblePlaybookCmd(
		playbook.WithPlaybooks(plbs...),
		playbook.WithPlaybookOptions(ansiblePlaybookOptions),
	)
	exec := measure.NewExecutorTimeMeasurement(
		execute.NewDefaultExecute(
			execute.WithCmd(playbookTasks),
		),
	)
	err = exec.Execute(context.TODO())
	if err != nil {
		return err
	}
	return nil
}
func newOpsTaskPlaybook(playbookKeys string) ([]string, error) {
	var plbs []string
	var playbook task.OpsPlaybook
	keys := strings.Split(playbookKeys, ",")
	for _, key := range keys {
		res, err := config.Etcd.Get(context.Background(), AddNamespace(key, os.Getenv("GAME_PLAYBOOK_NAMESPACE")))
		if err != nil {
			return nil, err
		}
		playbook.Key = string(res.Kvs[0].Key)
		playbook.Playbook = string(res.Kvs[0].Value)
		tempf, err := os.CreateTemp("/tmp", fmt.Sprintf("ops-task-%s-*.yml", key))
		if err != nil {
			return plbs, err
		}
		//defer tempf.Close()
		_, err = tempf.WriteString(playbook.Playbook)
		if err != nil {
			return plbs, nil
		}
		plbs = append(plbs, tempf.Name())
	}
	return plbs, nil
}
func newCommandPlaybook(s *datasource.SaurfangDatasources, ts *task.SaurfangPublishtasks, cnf map[string]serverconfig.Configs) (string, error) {
	var buf bytes.Buffer
	//var varsMap map[string]interface{} = make(map[string]interface{})
	var deployTask task.TemplateData
	for id, c := range cnf {
		host := task.Host{
			Host:       c.IP,
			BecomeUser: ts.BecomeUser,
		}

		if ts.Become == 1 {
			host.Become = "yes"
		} else {
			host.Become = "no"
		}
		host.Tasks = append(host.Tasks, task.Task{
			Name:      "Deploy" + "\t" + id + "\t" + c.SvcName,
			Prefix:    c.Prefix,
			Dest:      c.ConfigDir,
			EndPoint:  s.EndPoint,
			Region:    s.Region,
			Bucket:    s.Bucket,
			Path:      s.Path,
			Provider:  s.Provider,
			Profile:   s.Profile,
			AccessKey: s.AccessKey,
			SecretKey: s.SecretKey,
		})
		fmt.Println("tasks = ", host.Tasks)
		deployTask.Hosts = append(deployTask.Hosts, host)
	}
	fmt.Println("data = ", deployTask)
	tmpl := `
{{ range .Hosts }}
- name: "Tasks for host {{ .Host }}"
  gather_facts: no
  hosts: {{ .Host }}
  become: {{ .Become }}
  {{ if .BecomeUser }}
  become_user: {{ .BecomeUser }}
  {{ end }}
  tasks:
    {{ range .Tasks }}
    - name: check rclone config
      shell: rclone listremotes | grep {{ .Profile }}
      register: rclone_check
      ignore_errors: yes
    - name: "{{ .Name }}"
      ansible.builtin.shell: |
        rclone config create {{ .Profile }} s3 provider={{ .Provider }} access_key_id={{ .AccessKey }} secret_access_key={{ .SecretKey }} region={{ .Region }} endpoint={{ .EndPoint }}
      when: rclone_check.rc != 0
    - name: deploy server
      ansible.builtin.shell: |
        rclone sync  --local-no-set-modtime   --no-update-modtime --no-update-dir-modtime  {{ .Profile }}:{{ .Bucket }}{{ .Path }}{{ .Prefix }} {{ .Dest }}
    {{ end }}
{{ end }}
`
	t, err := template.New("cmd").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("template 错误:%v", err.Error())
	}
	err = t.Execute(&buf, deployTask)
	if err != nil {
		return "", err
	}
	fmt.Println(buf.String())
	timeStamp := time.Now().Format("20060102150405")
	tempf, err := os.CreateTemp("/tmp", fmt.Sprintf("publish-%s-*.yml", timeStamp))
	if err != nil {
		return "", err
	}
	_, err = io.WriteString(tempf, buf.String())
	if err != nil {
		return "", err
	}
	return tempf.Name(), nil
}
