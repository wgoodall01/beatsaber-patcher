package cmd

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/Jeffail/gabs"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "beatsaber-patcher",
	Short: "Patch Beat Saber maps with correct BPM information",
	Long: `Patch Beat Saber maps with correct BPM information. 
	
In the most recent update to Beat Saber (v0.10.2), support for certain
custom maps was broken. The cause of the issue was that BPM information
listed in the map data (e.g. Expert.json, etc) is now ignored, instead
reading BPM from the song manifest (info.json). This patcher copies
BPM from a the hardest difficulty level back into the song manifest,
so the game works properly with custom maps again
`,

	RunE: func(cmd *cobra.Command, args []string) error {
		var songsDir string
		if len(args) == 0 {
			// if no args passed, use enclosing directory of executable.
			binPath, binPathErr := os.Executable()
			if binPathErr != nil {
				return errors.Wrap(binPathErr, "can't find path to executable")
			}
			songsDir = filepath.Dir(binPath)
		} else {
			var pathErr error
			songsDir, pathErr = filepath.Abs(args[0])

			if pathErr != nil {
				return errors.Wrap(pathErr, "invalid path to songs directory")
			}
		}

		fileInfo, statErr := os.Stat(songsDir)
		if statErr != nil {
			return errors.Wrap(statErr, "couldn't open songs directory")
		}
		if !fileInfo.Mode().IsDir() {
			return errors.New("given path is not a directory")
		}

		infof("using beatsaber CustomSongs at %v", songsDir)

		// Get all the folders
		songsDirContents, lsErr := ioutil.ReadDir(songsDir)
		if lsErr != nil {
			return errors.Wrap(lsErr, "couldn't read contsnts of songs directory")
		}

		for _, songInfo := range songsDirContents {
			songDir := songInfo.Name()
			if !songInfo.Mode().IsDir() {
				// Not a directory, skip this.
				verbosef("not a directory, skipping: %s", songInfo.Name())
				continue
			}
			if strings.HasPrefix(songInfo.Name(), ".") {
				verbosef("is a dotfile, skipping: %s", songInfo.Name())
				continue
			}
			// Load the info.json manifest file, if it exists
			manifest, manifestErr := gabs.ParseJSONFile(filepath.Join(songsDir, songDir, "info.json"))
			if manifestErr != nil {
				warningf("couldn't parse manifest for: %s", songInfo.Name())
				continue
			}

			// Find manifest-level BPM
			manifestBpm := manifest.Path("beatsPerMinute").Data().(float64)

			// find the difficulty levels
			difficultyLevels, diffErr := manifest.Path("difficultyLevels").Children()
			if diffErr != nil {
				warningf("no difficulty levels in manifest for: %s", songInfo.Name())
			}

			verbosef("%v: manifest bpm=%.1f", songInfo.Name(), manifestBpm)

			hasMismatch := false
			trackBpms := make(map[string]float64)

			for _, manifestTrack := range difficultyLevels {
				difficultyLevel := manifestTrack.Path("difficulty").Data().(string)
				trackPath := manifestTrack.Path("jsonPath").Data().(string)

				track, trackErr := gabs.ParseJSONFile(filepath.Join(songsDir, songDir, trackPath))
				if trackErr != nil {
					warningf("couldn't parse track file: %s", filepath.Join(songDir, trackPath))
				}

				trackBpm := track.Path("_beatsPerMinute").Data().(float64)
				trackBpms[difficultyLevel] = trackBpm

				matchesManifest := closeEnough(trackBpm, manifestBpm)

				verbosef("%s: %s bpm=%.1f (match=%t)",
					songInfo.Name(),
					difficultyLevel,
					trackBpm,
					matchesManifest,
				)

				if !matchesManifest {
					infof(
						"%s: mismatched BPM: %s (%.1f) and manifest (%1f)",
						songInfo.Name(),
						difficultyLevel,
						trackBpm,
						manifestBpm,
					)
					hasMismatch = true
				}
			}

			if !DryRun && hasMismatch {
				var patchBpm float64

				useFrom := func(d string) {
					if trackBpms[d] != 0 {
						patchBpm = trackBpms[d]
					}
				}

				// Use the BPM of the hardest track
				useFrom("Easy")
				useFrom("Normal")
				useFrom("Hard")
				useFrom("Expert")
				useFrom("ExpertPlus")

				// Patch the info.json with the new BPM
				manifest.Set(manifestBpm, "beatsaber-patcher_bpm_old")
				manifest.Set(patchBpm, "beatsPerMinute")
				ioutil.WriteFile(filepath.Join(songsDir, songDir, "info.json"), manifest.EncodeJSON(), 0644)
				infof("%s: patched info.json to %.1f bpm", songInfo.Name(), patchBpm)

			}

		}

		return nil
	},

	Args: cobra.MaximumNArgs(1),
}

func closeEnough(a, b float64) bool {
	return math.Abs(a-b) < 1e-4
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func info(args ...interface{}) {
	fmt.Println(args...)
}

func infof(msg string, args ...interface{}) {
	fmt.Printf(msg+"\n", args...)
}

func warning(args ...interface{}) {
	info(append([]interface{}{"WARNING: "}, args...))
}

func warningf(msg string, args ...interface{}) {
	infof("WARNING: "+msg, args...)
}

func verbose(args ...interface{}) {
	if Verbose {
		info(append([]interface{}{"WARNING: "}, args...))
	}
}

func verbosef(msg string, args ...interface{}) {
	if Verbose {
		infof(msg, args...)
	}
}

var Verbose bool
var DryRun bool

func init() {
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolVarP(&DryRun, "dry-run", "d", false, "Don't write any changes to custom songs.")
	rootCmd.Flags().BoolVarP(&Verbose, "verbose", "v", false, "Verbose output")
}
