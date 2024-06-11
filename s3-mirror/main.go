package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"

	"github.com/blang/semver/v4"
	"github.com/google/go-github/v47/github"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/urfave/cli/v2"
)

// BinaryConfig represents the configuration for a binary to be downloaded
type BinaryConfig struct {
	Name        string   `yaml:"name"`
	Versions    Versions `yaml:"versions"`
	Download    string   `yaml:"download"`
	Checksum    string   `yaml:"checksum"`
	Destination string   `yaml:"destination"`
	Arch        []string `yaml:"arch"`
	OS          []string `yaml:"os"`
}

// Versions contains versioning information for a binary
type Versions struct {
	GitHub      string `yaml:"github"`
	Semver      string `yaml:"semver"`
	Prereleases bool   `yaml:"prereleases"`
}

// Config represents the entire YAML configuration
type Config struct {
	Binaries []BinaryConfig `yaml:"binaries"`
}

func FileExistsInS3(client *minio.Client, bucket, key string) (bool, error) {
	_, err := client.StatObject(context.Background(), bucket, key, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// DownloadFile downloads a file from a URL and saves it to a destination
func DownloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file: %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// VerifyChecksum verifies the SHA256 checksum of a file against a given checksum URL
func VerifyChecksum(filePath, checksumURL string) error {
	resp, err := http.Get(checksumURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download checksum: %s", resp.Status)
	}

	expectedChecksum, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return err
	}

	actualChecksum := fmt.Sprintf("%x", hasher.Sum(nil))
	if strings.TrimSpace(string(expectedChecksum)) != actualChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", strings.TrimSpace(string(expectedChecksum)), actualChecksum)
	}

	return nil
}

// UploadFileToS3 uploads a file to an S3 bucket
func UploadFileToS3(client *minio.Client, bucket, key, filename string) error {
	_, err := client.FPutObject(context.Background(), bucket, key, filename, minio.PutObjectOptions{})
	return err
}

func fetchGitHubVersions(owner, repo string, includePrereleases bool) ([]string, error) {
	client := github.NewClient(nil)
	releases, _, err := client.Repositories.ListReleases(context.Background(), owner, repo, nil)
	if err != nil {
		return nil, err
	}

	var versions []string
	for _, release := range releases {
		if release.TagName != nil {
			if !includePrereleases && release.GetPrerelease() {
				continue
			}
			versions = append(versions, *release.TagName)
		}
	}

	return versions, nil
}

func filterVersions(versions []string, semverConstraint string) ([]string, error) {
	constraint, err := semver.ParseRange(semverConstraint)
	if err != nil {
		return nil, err
	}

	var filteredVersions []string
	for _, version := range versions {
		v, err := semver.ParseTolerant(version)
		if err != nil {
			continue
		}

		if constraint(v) {
			filteredVersions = append(filteredVersions, version) // Use the original version string
		}
	}

	return filteredVersions, nil
}

func main() {
	app := &cli.App{
		Name:  "k8s-binary-downloader",
		Usage: "Download Kubernetes release binaries and upload them to an S3 bucket",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "bucket",
				Usage:    "Bucket Name",
				Required: true,
				EnvVars:  []string{"S3_BUCKET"},
			},
			&cli.StringFlag{
				Name:     "config",
				Usage:    "YAML config file for binaries to download",
				Required: false,
				Value:    "config.yaml",
				EnvVars:  []string{"CONFIG_FILE"},
			},
			&cli.StringFlag{
				Name:     "region",
				Usage:    "Bucket region",
				Value:    "us-east-1",
				Required: false,
				EnvVars:  []string{"AWS_REGION"},
			},
			&cli.StringFlag{
				Name:     "access-key",
				Usage:    "AWS access key ID",
				Required: true,
				EnvVars:  []string{"AWS_ACCESS_KEY_ID"},
			},
			&cli.StringFlag{
				Name:     "secret-key",
				Usage:    "AWS secret access key",
				Required: true,
				EnvVars:  []string{"AWS_SECRET_ACCESS_KEY"},
			},
			&cli.StringFlag{
				Name:     "endpoint",
				Usage:    "S3 endpoint",
				Required: true,
				EnvVars:  []string{"S3_ENDPOINT"},
			},
			&cli.BoolFlag{
				Name:     "tlssecure",
				Usage:    "Use TLS for the S3 endpoint",
				Required: false,
				EnvVars:  []string{"S3_TLSSECURE"},
				Value:    true,
			},
		},
		Action: func(c *cli.Context) error {
			bucket := c.String("bucket")
			configFile := c.String("config")
			endpoint := c.String("endpoint")

			// Initialize MinIO client
			minioClient, err := minio.New(endpoint, &minio.Options{
				Creds:  credentials.NewStaticV4(c.String("access-key"), c.String("secret-key"), ""),
				Secure: c.Bool("tlssecure"),
				Region: c.String("region"),
			})
			if err != nil {
				return fmt.Errorf("failed to initialize MinIO client: %v", err)
			}

			// Read and parse the YAML config file
			configData, err := os.ReadFile(configFile)
			if err != nil {
				return fmt.Errorf("failed to read config file: %v", err)
			}

			var config Config
			if err := yaml.Unmarshal(configData, &config); err != nil {
				return fmt.Errorf("failed to parse config file: %v", err)
			}

			for _, binary := range config.Binaries {
				// Fetch GitHub versions
				ownerRepo := strings.Split(strings.TrimPrefix(binary.Versions.GitHub, "https://github.com/"), "/")
				if len(ownerRepo) != 2 {
					return fmt.Errorf("invalid GitHub URL: %s", binary.Versions.GitHub)
				}

				pre := false
				if binary.Versions.Prereleases {
					pre = true
				}

				versions, err := fetchGitHubVersions(ownerRepo[0], ownerRepo[1], pre)
				if err != nil {
					return fmt.Errorf("failed to fetch GitHub versions: %v", err)
				}

				// Filter versions based on semver constraint
				filteredVersions, err := filterVersions(versions, binary.Versions.Semver)
				if err != nil {
					return fmt.Errorf("failed to filter versions: %v", err)
				}

				fmt.Printf("Filtered versions for %s: %v\n", binary.Name, filteredVersions)

				for _, version := range filteredVersions {
					for _, system := range binary.OS {
						for _, arch := range binary.Arch {
							tmplContext := map[string]string{
								"name":    binary.Name,
								"version": version,
								"os":      system,
								"arch":    arch,
							}

							tmpl, err := template.New("download").Parse(binary.Download)
							if err != nil {
								return fmt.Errorf("failed to parse download template: %v", err)
							}

							var downloadURL bytes.Buffer
							err = tmpl.Execute(&downloadURL, tmplContext)
							if err != nil {
								return fmt.Errorf("failed to execute download template: %v", err)
							}

							tmpl, err = template.New("checksum").Parse(binary.Checksum)
							if err != nil {
								return fmt.Errorf("failed to parse checksum template: %v", err)
							}

							var checksumURL bytes.Buffer
							err = tmpl.Execute(&checksumURL, tmplContext)
							if err != nil {
								return fmt.Errorf("failed to execute checksum template: %v", err)
							}

							tmpl, err = template.New("destination").Parse(binary.Destination)
							if err != nil {
								return fmt.Errorf("failed to parse destination template: %v", err)
							}

							var s3Dest bytes.Buffer
							err = tmpl.Execute(&s3Dest, tmplContext)
							if err != nil {
								return fmt.Errorf("failed to execute destination template: %v", err)
							}

							// Check if file exists in S3
							key := s3Dest.String()
							exists, err := FileExistsInS3(minioClient, bucket, key)
							if err != nil {
								return fmt.Errorf("failed to check if file exists in S3: %v", err)
							}
							if exists {
								fmt.Printf("File already exists in s3://%s/%s, skipping download\n", bucket, key)
								continue
							}

							tmpFile, err := os.CreateTemp("", "binary-*")
							if err != nil {
								return fmt.Errorf("failed to create temporary file: %v", err)
							}
							// Always Remove downloaded file
							defer func() {
								tmpFile.Close()
								os.Remove(tmpFile.Name())
							}()

							fmt.Printf("Downloading %s from %s\n", binary.Name, downloadURL.String())
							if err := DownloadFile(downloadURL.String(), tmpFile.Name()); err != nil {
								fmt.Printf("failed to download file: %v", err)
								continue
							}

							fmt.Printf("Verifying checksum for %s\n", tmpFile.Name())
							if err := VerifyChecksum(tmpFile.Name(), checksumURL.String()); err != nil {
								fmt.Printf("checksum verification failed: %v", err)
								continue

							}

							fmt.Printf("Uploading %s to s3://%s/%s\n", tmpFile.Name(), bucket, key)
							if err := UploadFileToS3(minioClient, bucket, key, tmpFile.Name()); err != nil {
								return fmt.Errorf("failed to upload file to S3: %v", err)
							}
						}
					}
				}
			}

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
