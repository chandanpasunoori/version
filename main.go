package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Version struct {
	Major, Minor, Patch int
}

type SemVerList []Version

func (s SemVerList) Len() int {
	return len(s)
}

func (s SemVerList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SemVerList) Less(i, j int) bool {
	if s[i].Major != s[j].Major {
		return s[i].Major < s[j].Major
	}
	if s[i].Minor != s[j].Minor {
		return s[i].Minor < s[j].Minor
	}
	return s[i].Patch < s[j].Patch
}

var (
	moduleName  string
	releaseType string
)

// Function to parse the current version from the version file
func getCurrentModules() ([]string, []string, error) {
	cmd := exec.Command("git", "tag", "--list", "--sort=-v:refname")
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

	// Return the latest version
	return modules, releases, nil
}

// Function to parse the current version from the version file
func parseCurrentVersion(moduleName, releaseType string) (Version, error) {
	cmd := exec.Command("git", "tag", "--list", "--sort=-v:refname")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("command", cmd.String()).Msg("error in the git command")
		return Version{Major: 0, Minor: 0, Patch: 0}, nil
	}

	tags := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(tags) == 0 {
		// No tags found
		return Version{Major: 0, Minor: 0, Patch: 0}, nil
	}

	// Extract versions and sort them in descending order
	var versions SemVerList

	re := regexp.MustCompile(fmt.Sprintf(`^(%s)/(%s)/v(\d+)\.(\d+)\.(\d+)$`, moduleName, releaseType))
	for _, tag := range tags {
		if matches := re.FindStringSubmatch(tag); len(matches) == 6 {
			major, err := strconv.Atoi(matches[3])
			if err != nil {
				log.Error().Msgf("invalid version parsing")
				os.Exit(1)
			}
			minor, err := strconv.Atoi(matches[4])
			if err != nil {
				log.Error().Msgf("invalid version parsing")
				os.Exit(1)
			}
			patch, err := strconv.Atoi(matches[5])
			if err != nil {
				log.Error().Msgf("invalid version parsing")
				os.Exit(1)
			}
			versions = append(versions, Version{Major: major, Minor: minor, Patch: patch})
		}
	}

	if len(versions) == 0 {
		// No valid version tags found
		return Version{Major: 0, Minor: 0, Patch: 0}, nil
	}

	sort.Sort(sort.Reverse(versions))

	// Return the latest version
	return versions[0], nil
}

// Function to generate the next version based on the specified pattern
func generateNextVersion(moduleName, releaseType string, currentVersion Version) string {
	// Increment the patch version
	nextVersion := currentVersion
	nextVersion.Patch += 1
	if nextVersion.Patch > 9 {
		nextVersion.Minor += 1
		nextVersion.Patch = 0
	}
	if nextVersion.Minor > 9 {
		nextVersion.Major += 1
		nextVersion.Minor = 0
	}
	// Construct the next version
	return fmt.Sprintf("%s/%s/v%d.%d.%d", moduleName, releaseType, nextVersion.Major, nextVersion.Minor, nextVersion.Patch)
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

	flag.StringVar(&moduleName, "m", "", "module name")
	flag.StringVar(&releaseType, "r", "", "release type")
	flag.Parse()

	log.Info().Msg("Welcome to the Tag Generator CLI")

	modules, releases, err := getCurrentModules()
	if err != nil {
		log.Error().Err(err).Msgf("Error reading current modules: %v", err)
		return
	}

	if len(moduleName) == 0 {
		// Get input for module name
		log.Info().Strs("modules", modules).Msg("Enter module name from list:")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		moduleName = scanner.Text()

		if !slices.Contains(modules, moduleName) {
			log.Info().Msg("Are you sure you want to create new module (yes/no)?")
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			yesOrNo := scanner.Text()
			if yesOrNo != "yes" {
				log.Error().Msgf("invalid module selected")
				os.Exit(1)
				return
			}
		}
	}

	if len(releaseType) == 0 {
		// Get input for release type
		log.Info().Strs("releases", releases).Msg("Enter release type from list:")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		releaseType = scanner.Text()

		if !slices.Contains(releases, releaseType) {
			log.Info().Msg("Are you sure you want to create new release channel (yes/no)?")
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			yesOrNo := scanner.Text()
			if yesOrNo != "yes" {
				log.Error().Msgf("invalid release channel selected")
				os.Exit(1)
				return
			}
		}
	}

	if strings.ContainsRune(moduleName, ' ') || strings.ContainsRune(releaseType, ' ') {
		log.Error().Msgf("invalid characters (space) in moduleName or releaseType")
		os.Exit(1)
		return
	}

	multiRelease := []string{releaseType}
	if strings.ContainsRune(releaseType, ',') {
		multiRelease = strings.Split(releaseType, ",")
	}
	for _, r := range multiRelease {
		// Read and display the current version
		currentVersion, err := parseCurrentVersion(moduleName, r)
		if err != nil {
			log.Error().Err(err).Msgf("Error reading current version: %v", err)
			return
		}
		log.Info().Interface("version", currentVersion).Msgf("Current version")

		// Generate and display the next version
		nextVersion := generateNextVersion(moduleName, r, currentVersion)
		if nextVersion == "" {
			log.Error().Msg("Error generating next version. Exiting.")
			return
		}

		log.Info().Msgf("Generated next version: %s", nextVersion)

		if err = createGitTag(nextVersion); err != nil {
			log.Error().Msg("Error creating git tag. Exiting.")
			return
		}
	}

	log.Info().Msg("Tags updated in repository.")
}
