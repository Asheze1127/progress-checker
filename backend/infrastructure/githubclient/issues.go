package githubclient

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	githubissuecreator "github.com/Asheze1127/progress-checker/backend/application/service/github_issue_creator"
	"github.com/google/go-github/v84/github"
)

// Compile-time check that Client implements GitHubIssueCreator.
var _ githubissuecreator.GitHubIssueCreator = (*Client)(nil)

type Config struct {
	Token             string
	BaseURL           string
	AppIssuer         string
	AppInstallationID int64
	AppPrivateKeyPEM  string
}

type Client struct {
	httpClient *http.Client
	baseURL    string
	token      string
	appAuth    *appAuthConfig
}

type appAuthConfig struct {
	issuer         string
	installationID int64
	privateKey     *rsa.PrivateKey
}

func NewClient(config Config) (*Client, error) {
	return newClient(config, nil)
}

func newClient(config Config, httpClient *http.Client) (*Client, error) {
	baseURL, err := normalizeBaseURL(config.BaseURL)
	if err != nil {
		return nil, err
	}

	token := strings.TrimSpace(config.Token)
	if token != "" {
		return &Client{
			httpClient: httpClient,
			baseURL:    baseURL,
			token:      token,
		}, nil
	}

	appIssuer := strings.TrimSpace(config.AppIssuer)
	if appIssuer == "" {
		return nil, fmt.Errorf("github token or github app issuer is required")
	}

	privateKey, err := parsePrivateKey(config.AppPrivateKeyPEM)
	if err != nil {
		return nil, err
	}

	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
		appAuth: &appAuthConfig{
			issuer:         appIssuer,
			installationID: config.AppInstallationID,
			privateKey:     privateKey,
		},
	}, nil
}

func (c *Client) CreateIssue(ctx context.Context, input githubissuecreator.CreateIssueInput) (*githubissuecreator.CreatedIssue, error) {
	if c == nil {
		return nil, fmt.Errorf("github client is not initialized")
	}

	owner := strings.TrimSpace(input.Owner)
	if owner == "" {
		return nil, fmt.Errorf("issue owner is required")
	}

	repo := strings.TrimSpace(input.Repo)
	if repo == "" {
		return nil, fmt.Errorf("issue repo is required")
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, fmt.Errorf("issue title is required")
	}

	issueRequest := &github.IssueRequest{
		Title: github.Ptr(title),
	}

	if input.Body != "" {
		issueRequest.Body = github.Ptr(input.Body)
	}

	if len(input.Labels) > 0 {
		labels := sanitizeLabels(input.Labels)
		if len(labels) > 0 {
			issueRequest.Labels = &labels
		}
	}

	githubClient, err := c.issueClient(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	issue, _, err := githubClient.Issues.Create(ctx, owner, repo, issueRequest)
	if err != nil {
		return nil, fmt.Errorf("create github issue: %w", err)
	}

	return &githubissuecreator.CreatedIssue{
		Number: issue.GetNumber(),
		URL:    issue.GetHTMLURL(),
	}, nil
}

func sanitizeLabels(labels []string) []string {
	sanitizedLabels := make([]string, 0, len(labels))

	for _, label := range labels {
		trimmedLabel := strings.TrimSpace(label)
		if trimmedLabel == "" {
			continue
		}

		sanitizedLabels = append(sanitizedLabels, trimmedLabel)
	}

	return sanitizedLabels
}

func (c *Client) issueClient(ctx context.Context, owner string, repo string) (*github.Client, error) {
	if c.token != "" {
		return c.newGitHubClient(c.token)
	}

	if c.appAuth == nil {
		return nil, fmt.Errorf("github authentication is not configured")
	}

	appJWT, err := createAppJWT(c.appAuth.issuer, c.appAuth.privateKey, time.Now())
	if err != nil {
		return nil, fmt.Errorf("create github app jwt: %w", err)
	}

	appClient, err := c.newGitHubClient(appJWT)
	if err != nil {
		return nil, err
	}

	installationID := c.appAuth.installationID
	if installationID == 0 {
		installation, _, installationErr := appClient.Apps.FindRepositoryInstallation(ctx, owner, repo)
		if installationErr != nil {
			return nil, fmt.Errorf("find github app installation: %w", installationErr)
		}

		installationID = installation.GetID()
		if installationID == 0 {
			return nil, fmt.Errorf("github app installation id was empty")
		}
	}

	installationToken, _, err := appClient.Apps.CreateInstallationToken(ctx, installationID, nil)
	if err != nil {
		return nil, fmt.Errorf("create github app installation token: %w", err)
	}

	token := strings.TrimSpace(installationToken.GetToken())
	if token == "" {
		return nil, fmt.Errorf("github app installation token was empty")
	}

	return c.newGitHubClient(token)
}

func (c *Client) newGitHubClient(token string) (*github.Client, error) {
	githubClient := github.NewClient(c.httpClient).WithAuthToken(token)

	if c.baseURL == "" {
		return githubClient, nil
	}

	baseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse github base url: %w", err)
	}

	githubClient.BaseURL = baseURL

	return githubClient, nil
}

func normalizeBaseURL(rawURL string) (string, error) {
	trimmedURL := strings.TrimSpace(rawURL)
	if trimmedURL == "" {
		return "", nil
	}

	baseURL, err := url.Parse(trimmedURL)
	if err != nil {
		return "", fmt.Errorf("parse github base url: %w", err)
	}

	if baseURL.Scheme != "https" {
		return "", fmt.Errorf("github base url must use https scheme, got %q", baseURL.Scheme)
	}

	if !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path += "/"
	}

	return baseURL.String(), nil
}

func parsePrivateKey(privateKeyPEM string) (*rsa.PrivateKey, error) {
	normalizedPrivateKeyPEM := strings.ReplaceAll(strings.TrimSpace(privateKeyPEM), `\n`, "\n")

	block, _ := pem.Decode([]byte(normalizedPrivateKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("decode github app private key pem")
	}

	if privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return privateKey, nil
	}

	pkcs8Key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse github app private key: %w", err)
	}

	privateKey, ok := pkcs8Key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("github app private key must be rsa")
	}

	return privateKey, nil
}

func createAppJWT(issuer string, privateKey *rsa.PrivateKey, now time.Time) (string, error) {
	headerSegment, err := encodeJWTPart(map[string]string{
		"alg": "RS256",
		"typ": "JWT",
	})
	if err != nil {
		return "", err
	}

	payloadSegment, err := encodeJWTPart(map[string]any{
		"iat": now.Add(-1 * time.Minute).Unix(),
		"exp": now.Add(9 * time.Minute).Unix(),
		"iss": issuer,
	})
	if err != nil {
		return "", err
	}

	unsignedToken := headerSegment + "." + payloadSegment
	tokenHash := sha256.Sum256([]byte(unsignedToken))

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, tokenHash[:])
	if err != nil {
		return "", fmt.Errorf("sign jwt: %w", err)
	}

	return unsignedToken + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func encodeJWTPart(payload any) (string, error) {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal jwt payload: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(jsonPayload), nil
}
