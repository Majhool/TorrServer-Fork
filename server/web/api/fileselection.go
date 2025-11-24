package api

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"server/torr/state"
)

// isVideoFile checks if a file is a video file based on extension
func isVideoFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	videoExts := []string{
		".mp4", ".mkv", ".avi", ".mov", ".wmv", ".flv", ".webm",
		".m4v", ".mpg", ".mpeg", ".3gp", ".ogv", ".ts", ".m2ts",
	}
	for _, videoExt := range videoExts {
		if ext == videoExt {
			return true
		}
	}
	return false
}

// autoSelectFile performs automatic file selection based on provided parameters
// Returns the file ID or -1 if no suitable file found
func autoSelectFile(fileStats []*state.TorrentFileStat, filename, season, episode string) int {
	if len(fileStats) == 0 {
		return -1
	}

	// Case 1: filename is provided - search for exact or partial match
	if filename != "" {
		return selectByFilename(fileStats, filename)
	}

	// Case 2: season and episode are provided - search using regex patterns
	if season != "" && episode != "" {
		return selectBySeasonEpisode(fileStats, season, episode)
	}

	// Case 3: no parameters - select largest video file
	return selectLargestVideoFile(fileStats)
}

// selectByFilename searches for file by name (case-insensitive)
func selectByFilename(fileStats []*state.TorrentFileStat, filename string) int {
	filename = strings.ToLower(filename)

	// First try exact match
	for _, file := range fileStats {
		baseName := strings.ToLower(filepath.Base(file.Path))
		if baseName == filename {
			return file.Id
		}
	}

	// Then try partial match (contains)
	for _, file := range fileStats {
		pathLower := strings.ToLower(file.Path)
		if strings.Contains(pathLower, filename) && isVideoFile(file.Path) {
			return file.Id
		}
	}

	return -1
}

// selectBySeasonEpisode searches for file using season/episode patterns
func selectBySeasonEpisode(fileStats []*state.TorrentFileStat, season, episode string) int {
	// Pad season and episode with zeros if needed
	seasonNum := season
	episodeNum := episode
	if len(season) == 1 {
		seasonNum = "0" + season
	}
	if len(episode) == 1 {
		episodeNum = "0" + episode
	}

	// Build regex patterns
	patterns := []string{
		fmt.Sprintf(`[Ss]%s[Ee]%s`, seasonNum, episodeNum),           // S01E05
		fmt.Sprintf(`[Ss]%s\.?[Ee]%s`, seasonNum, episodeNum),        // S01.E05
		fmt.Sprintf(`%sx%s`, season, episodeNum),                      // 1x05
		fmt.Sprintf(`%sx%s`, seasonNum, episodeNum),                   // 01x05
		fmt.Sprintf(`[Ss]eason[.\s]?%s.*[Ee]pisode[.\s]?%s`, season, episode), // Season 1 Episode 5
	}

	var matches []*state.TorrentFileStat

	// Search for matches
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}

		for _, file := range fileStats {
			if !isVideoFile(file.Path) {
				continue
			}

			if re.MatchString(file.Path) {
				matches = append(matches, file)
			}
		}

		// If we found matches with this pattern, stop searching
		if len(matches) > 0 {
			break
		}
	}

	// If no matches found, return -1
	if len(matches) == 0 {
		return -1
	}

	// If multiple matches, select the largest one
	largestFile := matches[0]
	for _, file := range matches {
		if file.Length > largestFile.Length {
			largestFile = file
		}
	}

	return largestFile.Id
}

// selectLargestVideoFile selects the largest video file from the list
func selectLargestVideoFile(fileStats []*state.TorrentFileStat) int {
	var largestFile *state.TorrentFileStat

	for _, file := range fileStats {
		if !isVideoFile(file.Path) {
			continue
		}

		if largestFile == nil || file.Length > largestFile.Length {
			largestFile = file
		}
	}

	if largestFile == nil {
		return -1
	}

	return largestFile.Id
}
