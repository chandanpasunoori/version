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

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
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
	interactive    bool
	commitHash     string
)

// listModel is a simple list selection model for bubbletea
type listModel struct {
	choices  []string
	cursor   int
	selected string
	title    string
	done     bool
}

// multiSelectModel is a multi-selection list model for bubbletea
type multiSelectModel struct {
	choices     []string
	cursor      int
	selected    map[int]bool
	title       string
	done        bool
}

func (m listModel) Init() tea.Cmd {
	return nil
}

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "enter":
			if len(m.choices) > 0 {
				m.selected = m.choices[m.cursor]
				m.done = true
			}
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

func (m listModel) View() string {
	s := fmt.Sprintf("%s\n\n", m.title)

	if len(m.choices) == 0 {
		s += "No items available.\n\n"
		s += "Press 'q' to quit."
		return s
	}

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	s += "\nPress 'enter' to select, 'q' to quit."
	return s
}

func (m multiSelectModel) Init() tea.Cmd {
	return nil
}

func (m multiSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "enter":
			// Finalize selection
			m.done = true
			return m, tea.Quit
		case " ":
			// Toggle selection for current item
			if len(m.choices) > 0 {
				if m.selected == nil {
					m.selected = make(map[int]bool)
				}
				m.selected[m.cursor] = !m.selected[m.cursor]
			}
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

func (m multiSelectModel) View() string {
	s := fmt.Sprintf("%s\n\n", m.title)

	if len(m.choices) == 0 {
		s += "No items available.\n\n"
		s += "Press 'q' to quit."
		return s
	}

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checkbox := "[ ]"
		if m.selected != nil && m.selected[i] {
			checkbox = "[x]"
		}
		s += fmt.Sprintf("%s %s %s\n", cursor, checkbox, choice)
	}

	selectedCount := 0
	if m.selected != nil {
		for _, selected := range m.selected {
			if selected {
				selectedCount++
			}
		}
	}

	s += fmt.Sprintf("\nSelected: %d items", selectedCount)
	s += "\n\nUse space to select/deselect, enter to confirm, 'q' to quit."
	return s
}

// runInteractiveSelection runs an interactive list selection and returns the selected item
func runInteractiveSelection(title string, choices []string) (string, error) {
	if len(choices) == 0 {
		return "", fmt.Errorf("no choices available")
	}

	model := listModel{
		choices: choices,
		title:   title,
		cursor:  0,
	}

	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	if m, ok := finalModel.(listModel); ok && m.done {
		return m.selected, nil
	}

	return "", fmt.Errorf("no selection made")
}

// runInteractiveMultiSelection runs an interactive multi-selection and returns the selected items as comma-separated string
func runInteractiveMultiSelection(title string, choices []string) (string, error) {
	if len(choices) == 0 {
		return "", fmt.Errorf("no choices available")
	}

	model := multiSelectModel{
		choices:  choices,
		title:    title,
		cursor:   0,
		selected: make(map[int]bool),
	}

	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	if m, ok := finalModel.(multiSelectModel); ok && m.done {
		var selectedItems []string
		for i, selected := range m.selected {
			if selected && i < len(m.choices) {
				selectedItems = append(selectedItems, m.choices[i])
			}
		}
		if len(selectedItems) == 0 {
			return "", fmt.Errorf("no items selected")
		}
		return strings.Join(selectedItems, ","), nil
	}

	return "", fmt.Errorf("no selection made")
}

// getLastNCommits returns the last N commits with their hash and message
func getLastNCommits(n int) ([]string, []string, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return nil, nil, err
	}

	// Get the HEAD reference
	head, err := repo.Head()
	if err != nil {
		return nil, nil, err
	}

	// Get commit iterator
	commits, err := repo.Log(&git.LogOptions{From: head.Hash()})
	if err != nil {
		return nil, nil, err
	}

	var commitHashes []string
	var commitDisplays []string
	count := 0

	err = commits.ForEach(func(c *object.Commit) error {
		if count >= n {
			return fmt.Errorf("done") // Stop iteration
		}
		shortHash := c.Hash.String()[:7]
		message := strings.Split(c.Message, "\n")[0] // First line only
		if len(message) > 50 {
			message = message[:47] + "..."
		}
		commitHashes = append(commitHashes, c.Hash.String())
		commitDisplays = append(commitDisplays, fmt.Sprintf("%s - %s", shortHash, message))
		count++
		return nil
	})

	if err != nil && err.Error() != "done" {
		return nil, nil, err
	}

	return commitHashes, commitDisplays, nil
}

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
func createGitTag(tag string, commitHashStr string) error {
	// Open the git repository
	repo, err := git.PlainOpen(".")
	if err != nil {
		log.Error().Err(err).Str("tag", tag).Msg("Failed to open git repository")
		return err
	}

	var hash plumbing.Hash
	
	if commitHashStr == "" {
		// Default to HEAD if no commit hash specified
		head, err := repo.Head()
		if err != nil {
			log.Error().Err(err).Str("tag", tag).Msg("Failed to get HEAD reference")
			return err
		}
		hash = head.Hash()
	} else {
		// Resolve the commit hash (handles both short and full hashes)
		if len(commitHashStr) == 40 {
			// Full hash
			hash = plumbing.NewHash(commitHashStr)
		} else {
			// Short hash - need to resolve it
			resolved := false
			iter, err := repo.CommitObjects()
			if err != nil {
				log.Error().Err(err).Str("tag", tag).Msg("Failed to get commit objects")
				return err
			}
			
			err = iter.ForEach(func(c *object.Commit) error {
				if strings.HasPrefix(c.Hash.String(), commitHashStr) {
					hash = c.Hash
					resolved = true
					return fmt.Errorf("found") // Stop iteration
				}
				return nil
			})
			
			if !resolved {
				err := fmt.Errorf("commit not found: %s", commitHashStr)
				log.Error().Err(err).Str("tag", tag).Str("commit", commitHashStr).Msg("Failed to resolve commit hash")
				return err
			}
		}
		
		// Verify the commit exists
		_, err := repo.CommitObject(hash)
		if err != nil {
			log.Error().Err(err).Str("tag", tag).Str("commit", commitHashStr).Msg("Failed to find commit")
			return err
		}
	}

	// Create the tag
	_, err = repo.CreateTag(tag, hash, nil)
	if err != nil {
		log.Error().Err(err).Str("tag", tag).Msg("Git tag create error")
		return err
	}

	log.Info().Str("tag", tag).Str("commit", hash.String()[:7]).Msg("Git tag created successfully")
	return nil
}

func main() {

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"})
	zerolog.SetGlobalLevel(zerolog.TraceLevel)

	flag.StringVar(&moduleName, "m", "", "module name")
	flag.StringVar(&releaseChannel, "r", "", "release channel")
	flag.BoolVar(&interactive, "i", false, "enable interactive mode with bubbletea list selection")
	flag.StringVar(&commitHash, "c", "", "commit hash (short or full) to tag, defaults to HEAD if not specified")
	flag.Parse()

	log.Info().Msg("Welcome to the Tag Generator CLI")

	modules, releases, err := getCurrentModules()
	if err != nil {
		log.Error().Err(err).Msgf("Error reading current modules: %v", err)
		return
	}

	// Create a single scanner to avoid stdin buffer issues
	var inputScanner *bufio.Scanner

	if len(moduleName) == 0 {
		if interactive {
			// Interactive mode using bubbletea
			if len(modules) > 0 {
				selected, err := runInteractiveSelection("Select a module:", modules)
				if err != nil {
					log.Error().Err(err).Msg("Error in interactive module selection")
					os.Exit(1)
					return
				}
				moduleName = selected
			} else {
				// No existing modules, fallback to text input
				log.Info().Msg("No existing modules found. Please enter a new module name:")
				if inputScanner == nil {
					inputScanner = bufio.NewScanner(os.Stdin)
				}
				inputScanner.Scan()
				moduleName = inputScanner.Text()
			}
		} else {
			// Traditional text input mode
			log.Info().Strs("modules", modules).Msg("Enter module name from list:")
			if inputScanner == nil {
				inputScanner = bufio.NewScanner(os.Stdin)
			}
			inputScanner.Scan()
			moduleName = inputScanner.Text()

			if !slices.Contains(modules, moduleName) {
				log.Info().Msg("Are you sure you want to create new module (yes/no)?")
				inputScanner.Scan()
				yesOrNo := inputScanner.Text()
				if yesOrNo != "yes" {
					log.Error().Msgf("invalid module name entered")
					os.Exit(1)
					return
				}
			}
		}
	}

	// Handle commit selection in interactive mode
	if interactive && len(commitHash) == 0 {
		// Offer commit selection: current or list of last 5
		commitChoices := []string{"Current commit (HEAD)", "Select from last 5 commits"}
		commitChoice, err := runInteractiveSelection("Select commit to tag:", commitChoices)
		if err != nil {
			log.Error().Err(err).Msg("Error in commit selection")
			os.Exit(1)
			return
		}

		if commitChoice == "Select from last 5 commits" {
			hashes, displays, err := getLastNCommits(5)
			if err != nil {
				log.Error().Err(err).Msg("Error fetching commits")
				os.Exit(1)
				return
			}

			if len(displays) > 0 {
				selected, err := runInteractiveSelection("Select a commit:", displays)
				if err != nil {
					log.Error().Err(err).Msg("Error in commit selection")
					os.Exit(1)
					return
				}
				// Find the index of the selected display string
				for i, display := range displays {
					if display == selected {
						commitHash = hashes[i]
						break
					}
				}
			}
		}
		// If "Current commit (HEAD)" was selected, commitHash remains empty (default)
	}

	if len(releaseChannel) == 0 {
		if interactive {
			// Interactive mode using bubbletea
			if len(releases) > 0 {
				selected, err := runInteractiveMultiSelection("Select release channels (use space to select, enter to confirm):", releases)
				if err != nil {
					log.Error().Err(err).Msg("Error in interactive release channel selection")
					os.Exit(1)
					return
				}
				releaseChannel = selected
			} else {
				// No existing releases, fallback to text input
				log.Info().Msg("No existing release channels found. Please enter a new release channel:")
				if inputScanner == nil {
					inputScanner = bufio.NewScanner(os.Stdin)
				}
				inputScanner.Scan()
				releaseChannel = inputScanner.Text()
			}
		} else {
			// Traditional text input mode
			log.Info().Strs("releases", releases).Msg("Enter release channel from list:")
			if inputScanner == nil {
				inputScanner = bufio.NewScanner(os.Stdin)
			}
			inputScanner.Scan()
			releaseChannel = inputScanner.Text()

			if !slices.Contains(releases, releaseChannel) {
				log.Info().Msg("Are you sure you want to create new release channel (yes/no)?")
				inputScanner.Scan()
				yesOrNo := inputScanner.Text()
				if yesOrNo != "yes" {
					log.Error().Msgf("invalid release channel entered")
					os.Exit(1)
					return
				}
			}
		}
	}

	if strings.ContainsRune(moduleName, ' ') || strings.ContainsRune(releaseChannel, ' ') {
		log.Error().Msgf("invalid characters (space) in module name or release channel")
		os.Exit(1)
		return
	}

	if len(strings.TrimSpace(moduleName)) == 0 {
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

		if err = createGitTag(nextVersion, commitHash); err != nil {
			log.Error().Msg("Error creating git tag. Exiting.")
			return
		}
	}

	log.Info().Msg("Tags updated in local repository, 'git push --tags' and enjoy")
}
