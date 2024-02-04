package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Function to parse the current version from the version file
func getCurrentModules() ([]string, []string, error) {
	cmd := exec.Command("git", "tag", "--list", "--sort=-v:refname")
	fmt.Println(cmd.String())
	output, err := cmd.CombinedOutput()
	if err != nil {
		return []string{}, []string{}, err
	}

	tags := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(tags) == 0 {
		return []string{}, []string{}, err
	}

	moduleNameList := make(map[string]bool)
	releaseTypeList := make(map[string]bool)

	re := regexp.MustCompile(`^([a-z]+)/([a-z]+)/v(\d+\.\d+\.\d+)$`)
	for _, tag := range tags {
		if matches := re.FindStringSubmatch(tag); len(matches) == 4 {
			module, release := matches[1], matches[2]
			if _, ok := releaseTypeList[release]; !ok {
				releaseTypeList[release] = true
			}
			if _, ok := moduleNameList[module]; !ok {
				moduleNameList[module] = true
			}
		}
	}

	var modules []string
	for key := range moduleNameList {
		modules = append(modules, key)
	}
	var releases []string
	for key := range releaseTypeList {
		releases = append(releases, key)
	}

	fmt.Println(modules)
	fmt.Println(releases)

	// Return the latest version
	return modules, releases, nil
}

// Function to parse the current version from the version file
func parseCurrentVersion(moduleName, releaseType string) (string, error) {
	cmd := exec.Command("git", "tag", "--list", "--sort=-v:refname")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("command", cmd.String()).Msg("error in the git command")
		return "", err
	}

	tags := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(tags) == 0 {
		// No tags found
		return "0.0.0", nil
	}

	// Extract versions and sort them in descending order
	var versions []string

	re := regexp.MustCompile(fmt.Sprintf(`^(%s)/(%s)/v(\d+\.\d+\.\d+)$`, moduleName, releaseType))
	for _, tag := range tags {
		if matches := re.FindStringSubmatch(tag); len(matches) == 4 {
			version := matches[3]

			versions = append(versions, version)
		}
	}

	if len(versions) == 0 {
		// No valid version tags found
		return "0.0.0", nil
	}

	sort.Sort(sort.Reverse(sort.StringSlice(versions)))

	// Return the latest version
	return versions[0], nil
}

// Function to generate the next version based on the specified pattern
func generateNextVersion(moduleName, releaseType, currentVersion string) string {
	// Parse major, minor, and patch versions
	re := regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)
	matches := re.FindStringSubmatch(currentVersion)
	if len(matches) != 4 {
		return fmt.Sprintf("%s/%s/v%d.%d.%d", moduleName, releaseType, 0, 0, 1)
	}

	major, minor, patch := matches[1], matches[2], matches[3]

	// Increment the patch version
	nextPatch := fmt.Sprintf("%d", parseVersion(patch)+1)

	// Construct the next version
	nextVersion := fmt.Sprintf("%s/%s/v%s.%s.%s", moduleName, releaseType, major, minor, nextPatch)
	return nextVersion
}

// Function to parse the minor version from semver
func parseVersion(version string) int {
	var num int
	fmt.Sscanf(version, "%d", &num)
	return num
}

// Function to create a git tag
func createGitTag(tag string) error {
	cmd := exec.Command("git", "tag", tag)
	err := cmd.Run()
	if err != nil {
		log.Error().Err(err).Str("command", cmd.String()).Str("tag", tag).Msg("Git tag create error")
		return err
	}

	log.Info().Str("tag", tag).Msg("Git tag created successfully")
	return nil
}

func main() {

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"})
	zerolog.SetGlobalLevel(zerolog.TraceLevel)

	log.Info().Msg("Welcome to the Tag Generator CLI")

	modules, releases, err := getCurrentModules()
	if err != nil {
		log.Error().Err(err).Msgf("Error reading current modules: %v", err)
		return
	}

	// Get input for module name
	log.Info().Strs("modules", modules).Msg("Enter module name from list:")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	moduleName := scanner.Text()

	// Get input for release type
	log.Info().Strs("releases", releases).Msg("Enter release type from list:")
	scanner.Scan()
	releaseType := scanner.Text()

	// Read and display the current version
	currentVersion, err := parseCurrentVersion(moduleName, releaseType)
	if err != nil {
		log.Error().Err(err).Msgf("Error reading current version: %v", err)
		return
	}
	log.Info().Str("version", currentVersion).Msgf("Current version")

	// Validate release type
	if releaseType != "release" && releaseType != "staging" && releaseType != "canary" {
		log.Error().Msg("Invalid release type. Exiting.")
		return
	}

	// Generate and display the next version
	nextVersion := generateNextVersion(moduleName, releaseType, currentVersion)
	if nextVersion == "" {
		log.Error().Msg("Error generating next version. Exiting.")
		return
	}

	log.Info().Msgf("Generated next version: %s", nextVersion)

	if err = createGitTag(nextVersion); err != nil {
		log.Error().Msg("Error creating git tag. Exiting.")
		return
	}

	log.Info().Msg("Tags updated in repository.")
}
