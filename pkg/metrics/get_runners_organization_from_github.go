package metrics

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/spendesk/github-actions-exporter/pkg/config"

	"github.com/google/go-github/v45/github"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	runnersOrganizationGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "github_tr_runner_organization_current_status",
			Help: "runner status: 0: offline; 1: idle; 2: active; -1: unknown",
		},
		[]string{"organization", "os", "name", "id", "all_labels"},
	)
)

func getAllOrgRunners(orga string) []*github.Runner {
	var runners []*github.Runner
	opt := &github.ListOptions{PerPage: 200}

	for {
		resp, rr, err := client.Actions.ListOrganizationRunners(context.Background(), orga, opt)
		if rl_err, ok := err.(*github.RateLimitError); ok {
			log.Printf("ListOrganizationRunners ratelimited. Pausing until %s", rl_err.Rate.Reset.Time.String())
			time.Sleep(time.Until(rl_err.Rate.Reset.Time))
			continue
		} else if err != nil {
			log.Printf("ListOrganizationRunners error for org %s: %s", orga, err.Error())
			return runners
		}

		runners = append(runners, resp.Runners...)
		if rr.NextPage == 0 {
			break
		}
		opt.Page = rr.NextPage
	}
	return runners
}

func runnerLabelsToString(runner *github.Runner) string {
    var labels []string
    for _, label := range runner.Labels {
		labels = append(labels, label.GetName())
    }
    return strings.Join(labels, ",")
}

func runnerStatusToCode(runner *github.Runner) float64 {
	s := runner.GetStatus()
	if s == "offline" {
		return 0
	} else if s == "online" {
		if runner.GetBusy() {
			return 2
		} else {
			return 1
		}
	}
	log.Printf("Unknown runner status: '%s'. Returning -1", s)
	return -1
}

// getRunnersOrganizationFromGithub - return information about runners and their status for an organization
func getRunnersOrganizationFromGithub() {
	for {
		for _, orga := range config.Github.Organizations.Value() {
			runners := getAllOrgRunners(orga)
			for _, runner := range runners {
				labels_str := runnerLabelsToString(runner)
				runnersOrganizationGauge.WithLabelValues(orga, *runner.OS, *runner.Name, strconv.FormatInt(runner.GetID(), 10), labels_str).Set(runnerStatusToCode(runner))
			}
		}

		time.Sleep(time.Duration(config.Github.Refresh) * time.Second)
		runnersOrganizationGauge.Reset()
	}
}
