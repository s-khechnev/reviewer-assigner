// nolint
package main

import (
	"encoding/csv"
	"fmt"
	"os"
	prDomain "reviewer-assigner/internal/domain/pullrequests"
	"strconv"
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

const TeamCount = 20
const UserCount = 200
const PullRequestCount = 2000
const PullRequestReviewersCount = 3000

type Team struct {
	ID   int64
	Name string
}

func (t *Team) ToStringSlice() []string {
	return []string{strconv.FormatInt(t.ID, 10), t.Name}
}

type User struct {
	ID       int64
	UserID   string
	Name     string
	TeamID   int64
	IsActive bool
}

func (u *User) ToStringSlice() []string {
	return []string{
		strconv.FormatInt(u.ID, 10),
		u.UserID,
		u.Name,
		strconv.FormatInt(u.TeamID, 10),
		strconv.FormatBool(u.IsActive),
	}
}

type PullRequest struct {
	ID            int64
	PullRequestID string
	Name          string
	AuthorID      string
	Status        string
	CreatedAt     time.Time
	MergedAt      *time.Time
}

func (p *PullRequest) ToStringSlice() []string {
	mergedAt := ""
	if p.MergedAt != nil {
		mergedAt = p.MergedAt.Format(time.RFC3339)
	}
	return []string{
		strconv.FormatInt(p.ID, 10),
		p.PullRequestID,
		p.Name,
		p.AuthorID,
		p.Status,
		p.CreatedAt.Format(time.RFC3339),
		mergedAt,
	}
}

type PullRequestReviewer struct {
	PullRequestID int64
	ReviewerID    int64
}

func (p *PullRequestReviewer) ToStringSlice() []string {
	return []string{strconv.FormatInt(p.PullRequestID, 10), strconv.FormatInt(p.ReviewerID, 10)}
}

func generateStatusActivity() bool {
	return gofakeit.Number(1, 100) <= 90
}

func generateStatusPR() string {
	if gofakeit.Number(1, 100) <= 75 {
		return string(prDomain.StatusMerged)
	}
	return string(prDomain.StatusOpen)
}

func main() {
	teams := make([]Team, TeamCount)
	for i := range len(teams) {
		teams[i] = Team{
			ID:   int64(i + 1),
			Name: fmt.Sprintf("team-%d-%s", i, gofakeit.NounCollectivePeople()),
		}
	}

	users := make([]User, UserCount)
	for i := range len(users) {
		users[i] = User{
			ID:       int64(i + 1),
			UserID:   fmt.Sprintf("user-%s", gofakeit.ID()),
			Name:     gofakeit.FirstName() + " " + gofakeit.LastName(),
			TeamID:   int64(i%TeamCount + 1),
			IsActive: generateStatusActivity(),
		}
	}

	pullRequests := make([]PullRequest, PullRequestCount)
	for i := range len(pullRequests) {
		author := users[gofakeit.Number(1, UserCount-1)]
		status := generateStatusPR()
		createdAt := gofakeit.DateRange(time.Now().AddDate(0, -6, 0), time.Now())

		pr := PullRequest{
			ID:            int64(i + 1),
			PullRequestID: fmt.Sprintf("pr-%s", gofakeit.ID()),
			Name:          gofakeit.Sentence(),
			AuthorID:      author.UserID,
			Status:        status,
			CreatedAt:     createdAt,
		}

		if status == string(prDomain.StatusMerged) {
			range3weeks := time.Duration(gofakeit.Number(1, 30)) * 24 * time.Hour
			mergedAt := createdAt.Add(range3weeks)
			pr.MergedAt = &mergedAt
		}

		pullRequests[i] = pr
	}

	pullRequestReviewers := make([]PullRequestReviewer, 0, PullRequestReviewersCount)
	prReviewersMap := make(map[int64]map[int64]bool)

	activeUsers := make([]User, 0, len(users))
	for _, user := range users {
		if user.IsActive {
			activeUsers = append(activeUsers, user)
		}
	}

	for len(pullRequestReviewers) < PullRequestReviewersCount {
		pr := pullRequests[gofakeit.Number(0, len(pullRequests)-1)]

		if prReviewersMap[pr.ID] == nil {
			prReviewersMap[pr.ID] = make(map[int64]bool)
		}

		if len(prReviewersMap[pr.ID]) >= 2 {
			continue
		}

		var author User
		for _, u := range users {
			if u.UserID == pr.AuthorID {
				author = u
				break
			}
		}

		var reviewer User
		attempts := 0
		maxAttempts := 50
		for attempts < maxAttempts {
			reviewer = activeUsers[gofakeit.Number(0, len(activeUsers)-1)]

			if reviewer.ID != author.ID && !prReviewersMap[pr.ID][reviewer.ID] {
				break
			}
			attempts++
		}

		if attempts >= maxAttempts {
			continue
		}

		prReviewer := PullRequestReviewer{
			PullRequestID: pr.ID,
			ReviewerID:    reviewer.ID,
		}
		pullRequestReviewers = append(pullRequestReviewers, prReviewer)
		prReviewersMap[pr.ID][reviewer.ID] = true
	}

	teamRecords := make([][]string, len(teams))
	for i, team := range teams {
		teamRecords[i] = team.ToStringSlice()
	}

	userRecords := make([][]string, len(users))
	for i, user := range users {
		userRecords[i] = user.ToStringSlice()
	}

	pullRequestRecords := make([][]string, len(pullRequests))
	for i, pr := range pullRequests {
		pullRequestRecords[i] = pr.ToStringSlice()
	}

	pullRequestReviewerRecords := make([][]string, len(pullRequestReviewers))
	for i, prr := range pullRequestReviewers {
		pullRequestReviewerRecords[i] = prr.ToStringSlice()
	}

	files := []struct {
		filename string
		headers  []string
		records  [][]string
	}{
		{"teams.csv", []string{"id", "name"}, teamRecords},
		{"users.csv", []string{"id", "user_id", "name", "team_id", "is_active"}, userRecords},
		{
			"pull_requests.csv",
			[]string{
				"id",
				"pull_request_id",
				"name",
				"author_id",
				"status",
				"created_at",
				"merged_at",
			},
			pullRequestRecords,
		},
		{
			"pull_request_reviewers.csv",
			[]string{"pull_request_id", "reviewer_id"},
			pullRequestReviewerRecords,
		},
	}

	for _, file := range files {
		if err := writeToCSV(file.filename, file.headers, file.records); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("Data successfully written to %s\n", file.filename)
	}

	fmt.Println("All data successfully written to CSV files!")
}

func writeToCSV(filename string, headers []string, records [][]string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file %s: %w", filename, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	if err = writer.Write(headers); err != nil {
		return fmt.Errorf("error writing headers to %s: %w", filename, err)
	}

	for _, record := range records {
		if err = writer.Write(record); err != nil {
			return fmt.Errorf("error writing record to %s: %w", filename, err)
		}
	}

	writer.Flush()
	if err = writer.Error(); err != nil {
		return fmt.Errorf("error flushing writer for %s: %w", filename, err)
	}

	return nil
}
