package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
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
	moduleName     string
	releaseChannel string
)

// Function to parse the current version from the version file
func getCurrentModules() ([]string, []string, error) {
	// Open the git repository
	repo, err := git.PlainOpen(".")
	if err != nil {
		return []string{}, []string{}, err
	}

	// Get tag references
	tagRefs, err := repo.Tags()
	if err != nil {
		return []string{}, []string{}, err
	}

	var tags []string
	err = tagRefs.ForEach(func(ref *plumbing.Reference) error {
		tagName := ref.Name().Short()
		tags = append(tags, tagName)
		return nil
	})
	if err != nil {
		return []string{}, []string{}, err
	}

	if len(tags) == 0 {
		return []string{}, []string{}, nil
	}

	moduleNameList := make(map[string]bool)
	releaseChannelList := make(map[string]bool)

	re := regexp.MustCompile(`^([a-z]+)/([a-z]+)/v(\d+\.\d+\.\d+)$`)
	for _, tag := range tags {
		if matches := re.FindStringSubmatch(tag); len(matches) == 4 {
			module, release := matches[1], matches[2]
			if _, ok := releaseChannelList[release]; !ok {
				releaseChannelList[release] = true
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
	for key := range releaseChannelList {
		releases = append(releases, key)
	}

	// Return the latest version
	return modules, releases, nil
}

// Function to parse the current version from the version file
func parseCurrentVersion(moduleName string, releaseChannel []string) (Version, error) {
	// Open the git repository
	repo, err := git.PlainOpen(".")
	if err != nil {
		log.Error().Err(err).Msg("Failed to open git repository")
		return Version{Major: 0, Minor: 0, Patch: 0}, nil
	}

	// Get tag references
	tagRefs, err := repo.Tags()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get git tags")
		return Version{Major: 0, Minor: 0, Patch: 0}, nil
	}

	var tags []string
	err = tagRefs.ForEach(func(ref *plumbing.Reference) error {
		tagName := ref.Name().Short()
		tags = append(tags, tagName)
		return nil
	})
	if err != nil {
		log.Error().Err(err).Msg("Error iterating over tags")
		return Version{Major: 0, Minor: 0, Patch: 0}, nil
	}

	if len(tags) == 0 {
		// No tags found
		return Version{Major: 0, Minor: 0, Patch: 0}, nil
	}

	// Extract versions and sort them in descending order
	var versions SemVerList

	for _, rc := range releaseChannel {
		re := regexp.MustCompile(fmt.Sprintf(`^(%s)/(%s)/v(\d+)\.(\d+)\.(\d+)$`, moduleName, rc))
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
func generateNextVersion(moduleName, releaseChannel string, currentVersion Version) string {
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
	return fmt.Sprintf("%s/%s/v%d.%d.%d", moduleName, releaseChannel, nextVersion.Major, nextVersion.Minor, nextVersion.Patch)
}

// Function to create a git tag
func createGitTag(tag string) error {
	// Open the git repository
	repo, err := git.PlainOpen(".")
	if err != nil {
		log.Error().Err(err).Str("tag", tag).Msg("Failed to open git repository")
		return err
	}

	// Get the HEAD reference to find the current commit
	head, err := repo.Head()
	if err != nil {
		log.Error().Err(err).Str("tag", tag).Msg("Failed to get HEAD reference")
		return err
	}

	// Create the tag
	_, err = repo.CreateTag(tag, head.Hash(), nil)
	if err != nil {
		log.Error().Err(err).Str("tag", tag).Msg("Git tag create error")
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
	flag.StringVar(&releaseChannel, "r", "", "release channel")
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
				log.Error().Msgf("invalid module name entered")
				os.Exit(1)
				return
			}
		}
	}

	if len(releaseChannel) == 0 {
		// Get input for release channel
		log.Info().Strs("releases", releases).Msg("Enter release channel from list:")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		releaseChannel = scanner.Text()

		if !slices.Contains(releases, releaseChannel) {
			log.Info().Msg("Are you sure you want to create new release channel (yes/no)?")
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			yesOrNo := scanner.Text()
			if yesOrNo != "yes" {
				log.Error().Msgf("invalid release channel entered")
				os.Exit(1)
				return
			}
		}
	}

	if strings.ContainsRune(moduleName, ' ') || strings.ContainsRune(releaseChannel, ' ') {
		log.Error().Msgf("invalid characters (space) in module name or release channel")
		os.Exit(1)
		return
	}

	if len(strings.TrimSpace(releaseChannel)) == 0 {
		log.Error().Msgf("invalid module name entered")
		os.Exit(1)
	}

	if len(strings.TrimSpace(releaseChannel)) == 0 {
		log.Error().Msgf("invalid release channel entered")
		os.Exit(1)
	}

	multiRelease := []string{releaseChannel}
	if strings.ContainsRune(releaseChannel, ',') {
		log.Info().Msg("please note first release channel version will be used for all subsequent release channels")
		multiRelease = strings.Split(releaseChannel, ",")
	}

	// Read and display the current version
	currentVersion, err := parseCurrentVersion(moduleName, multiRelease)
	if err != nil {
		log.Error().Err(err).Msgf("Error reading current version: %v", err)
		return
	}

	log.Info().Interface("version", currentVersion).Msgf("Current version")
	for _, r := range multiRelease {
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

	log.Info().Msg("Tags updated in local repository, 'git push --tags' and enjoy")
}
