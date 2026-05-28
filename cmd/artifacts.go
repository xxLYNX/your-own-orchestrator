package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"yoo/internal/database"
	"yoo/internal/models"
	"yoo/internal/platform"

	"github.com/spf13/cobra"
)

var (
	artifactType        string
	artifactName        string
	artifactFile        string
	artifactURL         string
	artifactDescription string
	artifactRequired    bool
)

var artifactsCmd = &cobra.Command{
	Use:     "artifacts",
	Aliases: []string{"artifact", "art"},
	Short:   "Manage artifacts (inputs and outputs) for templated notes",
	Long: `Manage artifacts for templated notes.

Artifacts are inputs and outputs associated with a templated note. They can be:
- Files: Local file paths that will be stored as absolute paths
- URLs: Web links to resources
- Folders: Directory paths
- Text: Text content

Examples:
  yoo artifact list 42
  yoo artifact add 42 --type input --name source --file ./data.csv
  yoo artifact add 42 --type output --name report --url https://example.com/report
  yoo artifact show 42 source
  yoo artifact delete 42 source
  yoo artifact open 42 report`,
}

var artifactsListCmd = &cobra.Command{
	Use:   "list <note-id>",
	Short: "List all artifacts for a templated note",
	Long: `List all artifacts (inputs and outputs) for a templated note.

Artifacts are shown grouped by type (Inputs and Outputs), with indicators
for required artifacts and whether they have been provided.

Example:
  yoo artifact list 42`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runForTemplatedNoteArg(args, func(db *database.DB, ctx *database.TemplatedNoteContext) error {
			artifacts, err := database.ListArtifacts(db.Conn(), ctx.NoteTemplate.ID)
			if err != nil {
				return fmt.Errorf("failed to list artifacts: %w", err)
			}
			if len(artifacts) == 0 {
				fmt.Println("No artifacts defined for this note.")
				return nil
			}

			inputs, outputs := models.GroupArtifacts(artifacts)
			printArtifactGroup("Inputs:", inputs)
			printArtifactGroup("Outputs:", outputs)
			return nil
		})
	},
}

var artifactsAddCmd = &cobra.Command{
	Use:   "add <note-id>",
	Short: "Add an artifact to a templated note",
	Long: `Add an artifact (input or output) to a templated note.

You must specify either --file or --url to define the artifact value.
For files, the path will be validated and stored as an absolute path.

Examples:
  yoo artifact add 42 --type input --name source --file ./data.csv --description "Source data"
  yoo artifact add 42 --type output --name report --url https://example.com/report
  yoo artifact add 42 --type input --name config --file ./config.json --required`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if artifactType == "" {
			return fmt.Errorf("--type flag is required (input or output)")
		}
		if artifactType != "input" && artifactType != "output" {
			return fmt.Errorf("artifact type must be 'input' or 'output', got: %s", artifactType)
		}
		if artifactName == "" {
			return fmt.Errorf("--name flag is required")
		}
		if artifactFile == "" && artifactURL == "" {
			return fmt.Errorf("either --file or --url must be specified")
		}
		if artifactFile != "" && artifactURL != "" {
			return fmt.Errorf("cannot specify both --file and --url")
		}

		return runForTemplatedNoteArg(args, func(db *database.DB, ctx *database.TemplatedNoteContext) error {
			var artType, value string
			if artifactFile != "" {
				if _, err := os.Stat(artifactFile); os.IsNotExist(err) {
					return fmt.Errorf("file does not exist: %s", artifactFile)
				}
				absPath, err := filepath.Abs(artifactFile)
				if err != nil {
					return fmt.Errorf("failed to get absolute path: %w", err)
				}
				artType = "file"
				value = absPath
			} else {
				artType = "url"
				value = artifactURL
			}

			artifact := &models.Artifact{
				NoteTemplateID: ctx.NoteTemplate.ID,
				ArtifactType:   artifactType,
				Name:           artifactName,
				Type:           artType,
				Value:          value,
				Description:    artifactDescription,
				Required:       artifactRequired,
			}

			if err := database.CreateArtifact(db.Conn(), artifact); err != nil {
				return fmt.Errorf("failed to create artifact: %w", err)
			}

			fmt.Printf("✓ Added %s artifact '%s' (%s)\n", artifactType, artifactName, artType)
			if artifactFile != "" {
				fmt.Printf("  Path: %s\n", value)
			} else {
				fmt.Printf("  URL: %s\n", value)
			}
			return nil
		})
	},
}

var artifactsShowCmd = &cobra.Command{
	Use:   "show <note-id> <name>",
	Short: "Show artifact details",
	Long: `Show detailed information about a specific artifact.

Example:
  yoo artifact show 42 source`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[1]
		return runForTemplatedNoteArg(args[:1], func(db *database.DB, ctx *database.TemplatedNoteContext) error {
			artifact, err := database.GetArtifact(db.Conn(), ctx.NoteTemplate.ID, name)
			if err != nil {
				return err
			}

			fmt.Printf("Name: %s\n", artifact.Name)
			fmt.Printf("Artifact Type: %s\n", artifact.ArtifactType)
			fmt.Printf("Type: %s\n", artifact.Type)
			fmt.Printf("Value: %s\n", artifact.Value)
			if artifact.Description != "" {
				fmt.Printf("Description: %s\n", artifact.Description)
			}
			fmt.Printf("Required: %v\n", artifact.Required)
			fmt.Printf("Created: %s\n", artifact.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("Updated: %s\n", artifact.UpdatedAt.Format("2006-01-02 15:04:05"))

			if artifact.Type == "file" {
				if _, err := os.Stat(artifact.Value); os.IsNotExist(err) {
					fmt.Printf("\n⚠ Warning: File does not exist at path: %s\n", artifact.Value)
				} else {
					fmt.Printf("\n✓ File exists\n")
				}
			}
			return nil
		})
	},
}

var artifactsDeleteCmd = &cobra.Command{
	Use:   "delete <note-id> <name>",
	Short: "Delete an artifact",
	Long: `Delete an artifact from a templated note.

This only removes the artifact reference from the database,
it does not delete the actual file or URL.

Example:
  yoo artifact delete 42 source`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[1]
		return runForTemplatedNoteArg(args[:1], func(db *database.DB, ctx *database.TemplatedNoteContext) error {
			if err := database.DeleteArtifact(db.Conn(), ctx.NoteTemplate.ID, name); err != nil {
				return err
			}
			fmt.Printf("✓ Deleted artifact '%s'\n", name)
			return nil
		})
	},
}

var artifactsOpenCmd = &cobra.Command{
	Use:   "open <note-id> <name>",
	Short: "Open an artifact (file or URL)",
	Long: `Open an artifact with the default system application.

For files, this opens the file with the default application.
For URLs, this opens the URL in the default web browser.

Example:
  yoo artifact open 42 report`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[1]
		return runForTemplatedNoteArg(args[:1], func(db *database.DB, ctx *database.TemplatedNoteContext) error {
			artifact, err := database.GetArtifact(db.Conn(), ctx.NoteTemplate.ID, name)
			if err != nil {
				return err
			}
			if artifact.Value == "" {
				return fmt.Errorf("artifact '%s' has no value set", name)
			}
			if artifact.Type == "file" {
				if _, err := os.Stat(artifact.Value); os.IsNotExist(err) {
					return fmt.Errorf("file does not exist: %s", artifact.Value)
				}
			}
			if err := platform.OpenWithDefaultApp(artifact.Value); err != nil {
				return fmt.Errorf("failed to open artifact: %w", err)
			}
			fmt.Printf("✓ Opened %s: %s\n", artifact.Type, artifact.Value)
			return nil
		})
	},
}

func init() {
	rootCmd.AddCommand(artifactsCmd)

	artifactsCmd.AddCommand(artifactsListCmd)
	artifactsCmd.AddCommand(artifactsAddCmd)
	artifactsCmd.AddCommand(artifactsShowCmd)
	artifactsCmd.AddCommand(artifactsDeleteCmd)
	artifactsCmd.AddCommand(artifactsOpenCmd)

	artifactsAddCmd.Flags().StringVar(&artifactType, "type", "", "Artifact type (input or output)")
	artifactsAddCmd.Flags().StringVar(&artifactName, "name", "", "Artifact name")
	artifactsAddCmd.Flags().StringVar(&artifactFile, "file", "", "File path")
	artifactsAddCmd.Flags().StringVar(&artifactURL, "url", "", "URL")
	artifactsAddCmd.Flags().StringVar(&artifactDescription, "description", "", "Artifact description")
	artifactsAddCmd.Flags().BoolVar(&artifactRequired, "required", false, "Mark artifact as required")
}
